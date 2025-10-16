// ==============================================================================
// Nome: BureauTokenForm.tsx
// Descrição: Formulário para geração de tokens para Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  TextField,
  Button,
  Grid,
  FormControl,
  FormHelperText,
  CircularProgress,
  Alert,
  Divider,
  useTheme,
  Chip,
  FormControlLabel,
  Checkbox,
  IconButton,
  InputAdornment,
  Tooltip,
  Slider,
  Switch,
  FormLabel,
  FormGroup,
  Card,
  CardContent,
  Select,
  MenuItem
} from '@mui/material';
import {
  ContentCopy as CopyIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
  AccessTime as AccessTimeIcon,
  Security as SecurityIcon,
  Info as InfoIcon
} from '@mui/icons-material';
import { useForm, Controller } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { useTranslation } from '../../../hooks/useTranslation';
import { useMultiTenant } from '../../../hooks/useMultiTenant';
import { useQuery } from '@apollo/client';
import { GET_BUREAU_IDENTITY, GET_BUREAU_AUTORIZACOES } from '../graphql/bureauQueries';

// Esquema de validação com Yup
const tokenFormSchema = yup.object().shape({
  autorizacaoId: yup.string().required('Selecione uma autorização'),
  escopos: yup.array().of(yup.string()).min(1, 'Selecione pelo menos um escopo'),
  expiracaoMinutos: yup.number()
    .required('Campo obrigatório')
    .min(1, 'Mínimo de 1 minuto')
    .max(43200, 'Máximo de 30 dias (43200 minutos)'),
  ipRestrictions: yup.string().matches(/^$|^([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}(\/[0-9]{1,2})?)(,\s*[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}(\/[0-9]{1,2})?)*$/, 'Formato de IP inválido'),
  permitirRefresh: yup.boolean(),
  rotacaoAutomatica: yup.boolean(),
  usoUnico: yup.boolean()
});

interface BureauTokenFormProps {
  identityId: string;
  onSubmit: (data: {
    autorizacaoId: string;
    escopos: string[];
    expiracaoMinutos: number;
    permitirRefresh?: boolean;
    rotacaoAutomatica?: boolean;
    usoUnico?: boolean;
    ipRestrictions?: string;
  }) => Promise<any>;
  loading: boolean;
}

// Escopos disponíveis para tokens
const ESCOPOS_DISPONIVEIS = [
  { 
    valor: 'bureau:read:basic', 
    label: 'Consulta Básica',
    descricao: 'Permite consultar informações básicas de identificação no Bureau'
  },
  { 
    valor: 'bureau:read:score', 
    label: 'Consulta Score',
    descricao: 'Permite consultar o score de crédito do usuário'
  },
  { 
    valor: 'bureau:read:full', 
    label: 'Consulta Completa',
    descricao: 'Permite consultar o relatório completo de crédito'
  },
  { 
    valor: 'bureau:read:history', 
    label: 'Histórico de Consultas',
    descricao: 'Permite consultar o histórico de consultas do usuário'
  },
  { 
    valor: 'bureau:write:report', 
    label: 'Enviar Relatório',
    descricao: 'Permite enviar relatórios de pagamento para o Bureau'
  }
];

// Opções de tempo de expiração pré-definidas
const EXPIRACAO_OPCOES = [
  { valor: 15, label: '15 minutos' },
  { valor: 60, label: '1 hora' },
  { valor: 360, label: '6 horas' },
  { valor: 720, label: '12 horas' },
  { valor: 1440, label: '1 dia' },
  { valor: 10080, label: '1 semana' },
  { valor: 43200, label: '30 dias' }
];

/**
 * Componente de formulário para geração de tokens para Bureau de Créditos
 * 
 * Este componente permite:
 * - Selecionar uma autorização existente
 * - Escolher escopos de acesso
 * - Definir tempo de expiração do token
 * - Configurar opções avançadas (refresh, rotação, uso único, restrições IP)
 * - Visualizar o token gerado com opção de copiar
 * 
 * Implementa validação avançada e feedback visual de erros
 */
export const BureauTokenForm: React.FC<BureauTokenFormProps> = ({
  identityId,
  onSubmit,
  loading
}) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { currentTenant } = useMultiTenant();
  
  // Estados locais
  const [formError, setFormError] = useState<string | null>(null);
  const [formSuccess, setFormSuccess] = useState<boolean>(false);
  const [geradoToken, setGeradoToken] = useState<string | null>(null);
  const [geradoRefreshToken, setGeradoRefreshToken] = useState<string | null>(null);
  const [mostrarToken, setMostrarToken] = useState<boolean>(false);
  const [mostrarOpcoesAvancadas, setMostrarOpcoesAvancadas] = useState<boolean>(false);

  // Buscar autorizações válidas para o vínculo
  const { loading: loadingAutorizacoes, error: errorAutorizacoes, data: dataAutorizacoes } = useQuery(
    GET_BUREAU_AUTORIZACOES,
    {
      variables: { identityId },
      skip: !identityId
    }
  );

  const autorizacoes = dataAutorizacoes?.bureauCredito?.autorizacoesByIdentity || [];

  // Configuração do React Hook Form com validação Yup
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
    setValue
  } = useForm({
    resolver: yupResolver(tokenFormSchema),
    defaultValues: {
      autorizacaoId: '',
      escopos: [] as string[],
      expiracaoMinutos: 60,
      permitirRefresh: false,
      rotacaoAutomatica: false,
      usoUnico: false,
      ipRestrictions: ''
    }
  });

  const watchEscopos = watch('escopos');
  const watchPermitirRefresh = watch('permitirRefresh');

  // Função para aplicar um template de expiração pré-definido
  const aplicarTemplateExpiracao = (minutos: number) => {
    setValue('expiracaoMinutos', minutos);
  };

  // Função para copiar token para a área de transferência
  const copiarToken = async () => {
    if (geradoToken) {
      try {
        await navigator.clipboard.writeText(geradoToken);
        // Poderia mostrar um toast ou notificação de sucesso aqui
      } catch (err) {
        console.error('Erro ao copiar token:', err);
      }
    }
  };

  // Função para tratar envio do formulário
  const handleFormSubmit = async (data: any) => {
    try {
      setFormError(null);
      setGeradoToken(null);
      setGeradoRefreshToken(null);
      
      const submitData = {
        autorizacaoId: data.autorizacaoId,
        escopos: data.escopos,
        expiracaoMinutos: data.expiracaoMinutos,
        permitirRefresh: data.permitirRefresh,
        rotacaoAutomatica: data.rotacaoAutomatica,
        usoUnico: data.usoUnico
      };

      // Adicionar restrições de IP se fornecidas
      if (data.ipRestrictions) {
        submitData.ipRestrictions = data.ipRestrictions;
      }

      // Chamar função de submissão passada como prop
      const response = await onSubmit(submitData);
      
      if (response) {
        setFormSuccess(true);
        setGeradoToken(response.token);
        
        if (response.refreshToken) {
          setGeradoRefreshToken(response.refreshToken);
        }
        
        // Limpar mensagem de sucesso após alguns segundos
        setTimeout(() => {
          setFormSuccess(false);
        }, 8000);
      }
    } catch (error) {
      console.error('Erro ao processar formulário:', error);
      setFormError(t('bureau.erro_gerar_token'));
    }
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Paper sx={{ width: '100%', p: 3 }}>
        <Typography variant="h6" gutterBottom>
          {t('bureau.gerar_token_titulo')}
        </Typography>
        <Typography variant="body2" color="textSecondary" paragraph>
          {t('bureau.gerar_token_descricao')}
        </Typography>

        {/* Mensagem de erro */}
        {formError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {formError}
          </Alert>
        )}

        {/* Mensagem de sucesso */}
        {formSuccess && (
          <Alert severity="success" sx={{ mb: 2 }}>
            {t('bureau.token_gerado_sucesso')}
          </Alert>
        )}

        <form onSubmit={handleSubmit(handleFormSubmit)}>
          <Grid container spacing={3}>
            {/* Seleção de autorização */}
            <Grid item xs={12}>
              <Controller
                name="autorizacaoId"
                control={control}
                render={({ field }) => (
                  <FormControl fullWidth error={!!errors.autorizacaoId}>
                    <TextField
                      {...field}
                      select
                      label={t('bureau.selecionar_autorizacao')}
                      error={!!errors.autorizacaoId}
                      helperText={errors.autorizacaoId?.message || t('bureau.selecionar_autorizacao_helper')}
                      disabled={loading || loadingAutorizacoes}
                    >
                      {loadingAutorizacoes ? (
                        <MenuItem disabled>
                          <CircularProgress size={20} sx={{ mr: 1 }} />
                          {t('comum.carregando')}
                        </MenuItem>
                      ) : autorizacoes.length === 0 ? (
                        <MenuItem disabled>
                          {t('bureau.nenhuma_autorizacao')}
                        </MenuItem>
                      ) : (
                        autorizacoes.map((auth: any) => (
                          <MenuItem key={auth.id} value={auth.id}>
                            {auth.tipoConsulta} - {auth.finalidade} ({new Date(auth.dataExpiracao).toLocaleDateString()})
                          </MenuItem>
                        ))
                      )}
                    </TextField>
                    {errorAutorizacoes && (
                      <FormHelperText error>
                        {t('bureau.erro_carregar_autorizacoes')}
                      </FormHelperText>
                    )}
                  </FormControl>
                )}
              />
            </Grid>

            {/* Seleção de escopos */}
            <Grid item xs={12}>
              <FormControl fullWidth error={!!errors.escopos}>
                <FormLabel>{t('bureau.selecionar_escopos')}</FormLabel>
                <FormGroup>
                  {ESCOPOS_DISPONIVEIS.map((escopo) => (
                    <Controller
                      key={escopo.valor}
                      name="escopos"
                      control={control}
                      render={({ field }) => (
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={field.value.includes(escopo.valor)}
                              onChange={(e) => {
                                const newEscopos = e.target.checked
                                  ? [...field.value, escopo.valor]
                                  : field.value.filter((v: string) => v !== escopo.valor);
                                field.onChange(newEscopos);
                              }}
                              disabled={loading}
                            />
                          }
                          label={
                            <Box>
                              <Typography variant="body2">{escopo.label}</Typography>
                              <Typography variant="caption" color="textSecondary">
                                {escopo.descricao}
                              </Typography>
                            </Box>
                          }
                        />
                      )}
                    />
                  ))}
                </FormGroup>
                {errors.escopos && (
                  <FormHelperText error>
                    {errors.escopos.message}
                  </FormHelperText>
                )}
              </FormControl>

              {/* Exibição de escopos selecionados como chips */}
              {watchEscopos.length > 0 && (
                <Box sx={{ mt: 1, display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                  {watchEscopos.map((escopo: string) => {
                    const escopoInfo = ESCOPOS_DISPONIVEIS.find(e => e.valor === escopo);
                    return (
                      <Chip 
                        key={escopo} 
                        label={escopoInfo?.label || escopo} 
                        size="small"
                        color="primary"
                        onDelete={() => {
                          setValue(
                            'escopos', 
                            watchEscopos.filter((e: string) => e !== escopo)
                          );
                        }}
                      />
                    );
                  })}
                </Box>
              )}
            </Grid>

            {/* Tempo de expiração */}
            <Grid item xs={12}>
              <FormLabel>{t('bureau.expiracao_token')}</FormLabel>
              <Box sx={{ mt: 2, mb: 1 }}>
                {EXPIRACAO_OPCOES.map((opcao) => (
                  <Button
                    key={opcao.valor}
                    variant="outlined"
                    size="small"
                    onClick={() => aplicarTemplateExpiracao(opcao.valor)}
                    sx={{ mr: 1, mb: 1 }}
                    startIcon={<AccessTimeIcon />}
                  >
                    {opcao.label}
                  </Button>
                ))}
              </Box>

              <Controller
                name="expiracaoMinutos"
                control={control}
                render={({ field }) => (
                  <FormControl fullWidth error={!!errors.expiracaoMinutos}>
                    <Grid container spacing={2} alignItems="center">
                      <Grid item xs>
                        <Slider
                          {...field}
                          min={1}
                          max={43200}
                          step={1}
                          valueLabelDisplay="auto"
                          valueLabelFormat={(value) => {
                            if (value < 60) return `${value} min`;
                            if (value < 1440) return `${Math.round(value / 60)} h`;
                            return `${Math.round(value / 1440)} dias`;
                          }}
                          disabled={loading}
                        />
                      </Grid>
                      <Grid item xs={3}>
                        <TextField
                          {...field}
                          type="number"
                          label={t('bureau.minutos')}
                          error={!!errors.expiracaoMinutos}
                          InputProps={{
                            endAdornment: (
                              <InputAdornment position="end">
                                <Tooltip title={t('bureau.expiracao_info')}>
                                  <InfoIcon color="action" fontSize="small" />
                                </Tooltip>
                              </InputAdornment>
                            ),
                          }}
                          disabled={loading}
                          fullWidth
                        />
                      </Grid>
                    </Grid>
                    {errors.expiracaoMinutos && (
                      <FormHelperText error>
                        {errors.expiracaoMinutos.message}
                      </FormHelperText>
                    )}
                    <FormHelperText>
                      {field.value < 60 
                        ? `${field.value} minutos` 
                        : field.value < 1440 
                          ? `${Math.round(field.value / 60)} horas` 
                          : `${Math.round(field.value / 1440)} dias`}
                    </FormHelperText>
                  </FormControl>
                )}
              />
            </Grid>

            {/* Toggle para mostrar/esconder opções avançadas */}
            <Grid item xs={12}>
              <Button
                variant="text"
                onClick={() => setMostrarOpcoesAvancadas(!mostrarOpcoesAvancadas)}
                endIcon={
                  mostrarOpcoesAvancadas ? (
                    <VisibilityOffIcon fontSize="small" />
                  ) : (
                    <VisibilityIcon fontSize="small" />
                  )
                }
              >
                {mostrarOpcoesAvancadas 
                  ? t('comum.esconder_opcoes_avancadas')
                  : t('comum.mostrar_opcoes_avancadas')}
              </Button>
            </Grid>

            {/* Opções avançadas */}
            {mostrarOpcoesAvancadas && (
              <>
                <Grid item xs={12} sm={6}>
                  <Controller
                    name="permitirRefresh"
                    control={control}
                    render={({ field }) => (
                      <FormControlLabel
                        control={
                          <Switch
                            checked={field.value}
                            onChange={(e) => field.onChange(e.target.checked)}
                            disabled={loading}
                          />
                        }
                        label={t('bureau.permitir_refresh')}
                      />
                    )}
                  />
                  <FormHelperText>
                    {t('bureau.permitir_refresh_info')}
                  </FormHelperText>
                </Grid>

                <Grid item xs={12} sm={6}>
                  <Controller
                    name="rotacaoAutomatica"
                    control={control}
                    render={({ field }) => (
                      <FormControlLabel
                        control={
                          <Switch
                            checked={field.value}
                            onChange={(e) => field.onChange(e.target.checked)}
                            disabled={loading || !watchPermitirRefresh}
                          />
                        }
                        label={t('bureau.rotacao_automatica')}
                      />
                    )}
                  />
                  <FormHelperText>
                    {t('bureau.rotacao_automatica_info')}
                  </FormHelperText>
                </Grid>

                <Grid item xs={12} sm={6}>
                  <Controller
                    name="usoUnico"
                    control={control}
                    render={({ field }) => (
                      <FormControlLabel
                        control={
                          <Switch
                            checked={field.value}
                            onChange={(e) => field.onChange(e.target.checked)}
                            disabled={loading}
                          />
                        }
                        label={t('bureau.uso_unico')}
                      />
                    )}
                  />
                  <FormHelperText>
                    {t('bureau.uso_unico_info')}
                  </FormHelperText>
                </Grid>

                <Grid item xs={12}>
                  <Controller
                    name="ipRestrictions"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        fullWidth
                        label={t('bureau.restricao_ip')}
                        placeholder="192.168.1.1, 10.0.0.1/24"
                        error={!!errors.ipRestrictions}
                        helperText={
                          errors.ipRestrictions?.message ||
                          t('bureau.restricao_ip_info')
                        }
                        disabled={loading}
                        InputProps={{
                          startAdornment: (
                            <InputAdornment position="start">
                              <SecurityIcon color="action" />
                            </InputAdornment>
                          ),
                        }}
                      />
                    )}
                  />
                </Grid>
              </>
            )}

            {/* Botões de ação */}
            <Grid item xs={12}>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 2 }}>
                <Button
                  variant="outlined"
                  onClick={() => {
                    reset();
                    setGeradoToken(null);
                    setGeradoRefreshToken(null);
                  }}
                  disabled={loading}
                >
                  {t('comum.limpar')}
                </Button>
                <Button
                  type="submit"
                  variant="contained"
                  disabled={loading}
                  startIcon={loading && <CircularProgress size={20} />}
                >
                  {loading ? t('comum.processando') : t('bureau.gerar_token')}
                </Button>
              </Box>
            </Grid>
          </Grid>
        </form>

        {/* Exibição de token gerado */}
        {geradoToken && (
          <Box sx={{ mt: 3 }}>
            <Alert severity="info" icon={<SecurityIcon />}>
              {t('bureau.token_acesso_seguro')}
            </Alert>
            
            <Card variant="outlined" sx={{ mt: 2 }}>
              <CardContent>
                <Typography variant="subtitle2" gutterBottom>
                  {t('bureau.access_token')}:
                </Typography>
                <Box 
                  sx={{ 
                    position: 'relative',
                    display: 'flex', 
                    alignItems: 'center', 
                    backgroundColor: theme.palette.background.default,
                    p: 1,
                    borderRadius: 1,
                    mb: 2
                  }}
                >
                  <Typography
                    variant="body2"
                    sx={{
                      fontFamily: 'monospace',
                      overflow: 'auto',
                      flex: 1,
                      wordBreak: 'break-all'
                    }}
                  >
                    {mostrarToken ? geradoToken : '••••••••••••••••••••••••••••••••••••'}
                  </Typography>
                  <Box sx={{ display: 'flex', ml: 1 }}>
                    <IconButton 
                      size="small"
                      onClick={() => setMostrarToken(!mostrarToken)} 
                      aria-label="toggle token visibility"
                    >
                      {mostrarToken ? <VisibilityOffIcon /> : <VisibilityIcon />}
                    </IconButton>
                    <IconButton 
                      size="small" 
                      onClick={copiarToken} 
                      aria-label="copy token"
                    >
                      <CopyIcon />
                    </IconButton>
                  </Box>
                </Box>

                {geradoRefreshToken && (
                  <>
                    <Typography variant="subtitle2" gutterBottom>
                      {t('bureau.refresh_token')}:
                    </Typography>
                    <Box 
                      sx={{ 
                        position: 'relative',
                        display: 'flex', 
                        alignItems: 'center', 
                        backgroundColor: theme.palette.background.default,
                        p: 1,
                        borderRadius: 1
                      }}
                    >
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: 'monospace',
                          overflow: 'auto',
                          flex: 1,
                          wordBreak: 'break-all'
                        }}
                      >
                        {mostrarToken ? geradoRefreshToken : '••••••••••••••••••••••••••••••••••••'}
                      </Typography>
                    </Box>
                  </>
                )}
              </CardContent>
            </Card>
          </Box>
        )}
      </Paper>
    </Box>
  );
};