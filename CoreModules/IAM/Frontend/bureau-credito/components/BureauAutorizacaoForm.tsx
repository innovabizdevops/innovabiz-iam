// ==============================================================================
// Nome: BureauAutorizacaoForm.tsx
// Descrição: Formulário para criação de autorizações para Bureau de Créditos
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
  FormLabel,
  Autocomplete,
  Chip,
  RadioGroup,
  Radio,
  FormControlLabel
} from '@mui/material';
import {
  DatePicker,
  LocalizationProvider
} from '@mui/x-date-pickers';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { ptBR } from 'date-fns/locale';
import { addDays, differenceInDays } from 'date-fns';
import { useForm, Controller } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { useTranslation } from '../../../hooks/useTranslation';
import { useMultiTenant } from '../../../hooks/useMultiTenant';
import { useQuery } from '@apollo/client';
import { GET_BUREAU_IDENTITY } from '../graphql/bureauQueries';
import { TipoConsulta } from '../../../types/bureau-credito';

// Esquema de validação com Yup
const autorizacaoSchema = yup.object().shape({
  tipoConsulta: yup.string().oneOf(
    ['SIMPLES', 'COMPLETA', 'SCORE', 'ANALITICA'], 
    'Tipo de consulta inválido'
  ).required('Campo obrigatório'),
  finalidade: yup.string().required('Campo obrigatório').min(5, 'Descreva a finalidade com pelo menos 5 caracteres'),
  justificativa: yup.string().required('Campo obrigatório').min(10, 'Forneça uma justificativa detalhada'),
  diasValidade: yup.number()
    .required('Campo obrigatório')
    .min(1, 'Mínimo de 1 dia')
    .max(365, 'Máximo de 365 dias')
    .integer('Deve ser um número inteiro'),
  dataExpiracao: yup.date()
    .required('Campo obrigatório')
    .min(
      addDays(new Date(), 1), 
      'Data deve ser pelo menos 1 dia no futuro'
    )
    .max(
      addDays(new Date(), 365), 
      'Data não pode ser mais de 1 ano no futuro'
    ),
  tags: yup.array().of(yup.string()).min(1, 'Adicione pelo menos uma tag'),
  observacoes: yup.string(),
  confirmarUso: yup.boolean().oneOf([true], 'Você deve confirmar este termo')
});

interface BureauAutorizacaoFormProps {
  identityId: string;
  onSubmit: (data: {
    identityId: string;
    tipoConsulta: TipoConsulta;
    finalidade: string;
    justificativa: string;
    diasValidade: number;
    tags?: string[];
    observacoes?: string;
  }) => Promise<any>;
  loading: boolean;
}

// Sugestões de tags comuns para consultas
const TAGS_SUGERIDAS = [
  'Análise de Crédito',
  'Aprovação de Empréstimo',
  'Verificação de Identidade',
  'Abertura de Conta',
  'Gestão de Risco',
  'Compliance KYC',
  'Fraude',
  'Conformidade',
  'Monitoramento',
  'Score',
  'Verificação',
  'Avaliação',
  'Onboarding',
  'Periódica'
];

/**
 * Componente de formulário para criar autorizações para consultas ao Bureau de Créditos
 * 
 * Este componente permite:
 * - Selecionar o tipo de consulta a ser autorizada
 * - Definir a finalidade e justificativa da consulta
 * - Configurar o período de validade da autorização
 * - Adicionar tags para classificação e metadados
 * - Incluir observações complementares
 * 
 * Implementa validação avançada e feedback visual de erros
 */
export const BureauAutorizacaoForm: React.FC<BureauAutorizacaoFormProps> = ({
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
  const [submitResponse, setSubmitResponse] = useState<any>(null);

  // Buscar detalhes do vínculo para exibir informações
  const { loading: loadingIdentity, error: errorIdentity, data: identityData } = useQuery(
    GET_BUREAU_IDENTITY,
    {
      variables: { id: identityId },
      skip: !identityId
    }
  );

  const identity = identityData?.bureauCredito?.bureauIdentity || null;

  // Configuração do React Hook Form com validação Yup
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
    setValue,
    getValues
  } = useForm({
    resolver: yupResolver(autorizacaoSchema),
    defaultValues: {
      tipoConsulta: 'SIMPLES' as TipoConsulta,
      finalidade: '',
      justificativa: '',
      diasValidade: 30,
      dataExpiracao: addDays(new Date(), 30),
      tags: [] as string[],
      observacoes: '',
      confirmarUso: false
    }
  });

  // Observar a mudança da data de expiração para atualizar dias de validade
  const watchDataExpiracao = watch('dataExpiracao');

  // Atualizar diasValidade quando a dataExpiracao mudar
  React.useEffect(() => {
    if (watchDataExpiracao) {
      const dias = differenceInDays(watchDataExpiracao, new Date());
      if (dias >= 1 && dias <= 365) {
        setValue('diasValidade', dias);
      }
    }
  }, [watchDataExpiracao, setValue]);

  // Função para atualizar a data de expiração quando os dias de validade mudarem
  const handleDiasValidadeChange = (dias: number) => {
    setValue('diasValidade', dias);
    setValue('dataExpiracao', addDays(new Date(), dias));
  };

  // Função para tratar envio do formulário
  const handleFormSubmit = async (data: any) => {
    try {
      setFormError(null);
      setSubmitResponse(null);
      
      // Preparar dados para envio
      const submitData = {
        identityId,
        tipoConsulta: data.tipoConsulta,
        finalidade: data.finalidade,
        justificativa: data.justificativa,
        diasValidade: data.diasValidade,
        tags: data.tags,
        observacoes: data.observacoes
      };

      // Chamar função de submissão passada como prop
      const response = await onSubmit(submitData);
      
      if (response) {
        setFormSuccess(true);
        setSubmitResponse(response);
        
        // Resetar formulário após sucesso
        reset();
        
        // Limpar mensagem de sucesso após alguns segundos
        setTimeout(() => {
          setFormSuccess(false);
        }, 8000);
      }
    } catch (error) {
      console.error('Erro ao processar formulário:', error);
      setFormError(t('bureau.erro_criar_autorizacao'));
    }
  };

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns} adapterLocale={ptBR}>
      <Box sx={{ width: '100%' }}>
        <Paper sx={{ width: '100%', p: 3 }}>
          <Typography variant="h6" gutterBottom>
            {t('bureau.criar_autorizacao_titulo')}
          </Typography>
          <Typography variant="body2" color="textSecondary" paragraph>
            {t('bureau.criar_autorizacao_descricao')}
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
              {t('bureau.autorizacao_criada_sucesso')}
            </Alert>
          )}

          {/* Informações do vínculo */}
          {loadingIdentity ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', my: 2 }}>
              <CircularProgress size={24} />
            </Box>
          ) : errorIdentity ? (
            <Alert severity="error" sx={{ mb: 2 }}>
              {t('bureau.erro_carregar_vinculo')}
            </Alert>
          ) : identity ? (
            <Box sx={{ mb: 3, p: 2, bgcolor: theme.palette.background.default, borderRadius: 1 }}>
              <Grid container spacing={2}>
                <Grid item xs={12} md={6}>
                  <Typography variant="subtitle2">{t('bureau.usuario_vinculo')}</Typography>
                  <Typography variant="body2">{identity.usuarioNome}</Typography>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="subtitle2">{t('bureau.tipo_vinculo')}</Typography>
                  <Typography variant="body2">{t(`bureau.tipo_${identity.tipoVinculo.toLowerCase()}`)}</Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="subtitle2">{t('bureau.nivel_acesso')}</Typography>
                  <Typography variant="body2">{t(`bureau.nivel_${identity.nivelAcesso.toLowerCase()}`)}</Typography>
                </Grid>
              </Grid>
            </Box>
          ) : null}

          <form onSubmit={handleSubmit(handleFormSubmit)}>
            <Grid container spacing={3}>
              {/* Tipo de consulta */}
              <Grid item xs={12}>
                <Controller
                  name="tipoConsulta"
                  control={control}
                  render={({ field }) => (
                    <FormControl component="fieldset" error={!!errors.tipoConsulta}>
                      <FormLabel component="legend">{t('bureau.tipo_consulta')}</FormLabel>
                      <RadioGroup
                        row
                        {...field}
                      >
                        <FormControlLabel 
                          value="SIMPLES" 
                          control={<Radio />} 
                          label={t('bureau.consulta_simples')} 
                          disabled={loading}
                        />
                        <FormControlLabel 
                          value="COMPLETA" 
                          control={<Radio />} 
                          label={t('bureau.consulta_completa')} 
                          disabled={loading}
                        />
                        <FormControlLabel 
                          value="SCORE" 
                          control={<Radio />} 
                          label={t('bureau.consulta_score')} 
                          disabled={loading}
                        />
                        <FormControlLabel 
                          value="ANALITICA" 
                          control={<Radio />} 
                          label={t('bureau.consulta_analitica')} 
                          disabled={loading}
                        />
                      </RadioGroup>
                      {errors.tipoConsulta && (
                        <FormHelperText error>
                          {errors.tipoConsulta.message}
                        </FormHelperText>
                      )}
                    </FormControl>
                  )}
                />
              </Grid>

              {/* Finalidade */}
              <Grid item xs={12} md={6}>
                <Controller
                  name="finalidade"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t('bureau.finalidade')}
                      placeholder={t('bureau.finalidade_placeholder')}
                      error={!!errors.finalidade}
                      helperText={errors.finalidade?.message || t('bureau.finalidade_helper')}
                      disabled={loading}
                    />
                  )}
                />
              </Grid>

              {/* Tags */}
              <Grid item xs={12} md={6}>
                <Controller
                  name="tags"
                  control={control}
                  render={({ field }) => (
                    <Autocomplete
                      {...field}
                      multiple
                      freeSolo
                      options={TAGS_SUGERIDAS}
                      value={field.value}
                      onChange={(_, newValue) => field.onChange(newValue)}
                      renderTags={(value, getTagProps) =>
                        value.map((option, index) => (
                          <Chip
                            label={option}
                            size="small"
                            {...getTagProps({ index })}
                          />
                        ))
                      }
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label={t('bureau.tags')}
                          placeholder={t('bureau.tags_placeholder')}
                          error={!!errors.tags}
                          helperText={errors.tags?.message || t('bureau.tags_helper')}
                        />
                      )}
                      disabled={loading}
                    />
                  )}
                />
              </Grid>

              {/* Justificativa */}
              <Grid item xs={12}>
                <Controller
                  name="justificativa"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      multiline
                      rows={4}
                      label={t('bureau.justificativa')}
                      placeholder={t('bureau.justificativa_placeholder')}
                      error={!!errors.justificativa}
                      helperText={errors.justificativa?.message || t('bureau.justificativa_helper')}
                      disabled={loading}
                    />
                  )}
                />
              </Grid>

              {/* Dias de validade */}
              <Grid item xs={12} md={6}>
                <Controller
                  name="diasValidade"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label={t('bureau.dias_validade')}
                      error={!!errors.diasValidade}
                      helperText={errors.diasValidade?.message || t('bureau.dias_validade_helper')}
                      disabled={loading}
                      onChange={(e) => {
                        const valor = parseInt(e.target.value);
                        if (!isNaN(valor)) {
                          handleDiasValidadeChange(valor);
                        } else {
                          field.onChange(e);
                        }
                      }}
                      InputProps={{
                        inputProps: { min: 1, max: 365 }
                      }}
                    />
                  )}
                />
              </Grid>

              {/* Data de expiração */}
              <Grid item xs={12} md={6}>
                <Controller
                  name="dataExpiracao"
                  control={control}
                  render={({ field }) => (
                    <DatePicker
                      {...field}
                      label={t('bureau.data_expiracao')}
                      minDate={addDays(new Date(), 1)}
                      maxDate={addDays(new Date(), 365)}
                      disabled={loading}
                      slotProps={{
                        textField: {
                          fullWidth: true,
                          error: !!errors.dataExpiracao,
                          helperText: errors.dataExpiracao?.message || t('bureau.data_expiracao_helper')
                        }
                      }}
                    />
                  )}
                />
              </Grid>

              {/* Observações */}
              <Grid item xs={12}>
                <Controller
                  name="observacoes"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      multiline
                      rows={2}
                      label={t('bureau.observacoes')}
                      placeholder={t('bureau.observacoes_placeholder')}
                      error={!!errors.observacoes}
                      helperText={errors.observacoes?.message || t('bureau.observacoes_helper')}
                      disabled={loading}
                    />
                  )}
                />
              </Grid>

              {/* Confirmação */}
              <Grid item xs={12}>
                <Controller
                  name="confirmarUso"
                  control={control}
                  render={({ field }) => (
                    <FormControl error={!!errors.confirmarUso}>
                      <FormControlLabel
                        control={
                          <Radio
                            checked={field.value}
                            onChange={(e) => field.onChange(e.target.checked)}
                            disabled={loading}
                          />
                        }
                        label={t('bureau.confirmar_uso_legitimo')}
                      />
                      {errors.confirmarUso && (
                        <FormHelperText error>
                          {errors.confirmarUso.message}
                        </FormHelperText>
                      )}
                    </FormControl>
                  )}
                />
              </Grid>

              {/* Botões de ação */}
              <Grid item xs={12}>
                <Divider sx={{ mb: 2 }} />
                <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 2 }}>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      reset();
                    }}
                    disabled={loading}
                  >
                    {t('comum.cancelar')}
                  </Button>
                  <Button
                    type="submit"
                    variant="contained"
                    disabled={loading}
                    startIcon={loading && <CircularProgress size={20} />}
                  >
                    {loading ? t('comum.processando') : t('bureau.criar_autorizacao')}
                  </Button>
                </Box>
              </Grid>
            </Grid>
          </form>

          {/* Exibição da resposta */}
          {submitResponse && (
            <Box sx={{ mt: 3 }}>
              <Alert severity="info">
                <Typography variant="subtitle2">
                  {t('bureau.autorizacao_criada_id')}: {submitResponse.id}
                </Typography>
                <Typography variant="body2">
                  {t('bureau.autorizacao_valida_ate')}: {
                    new Date(submitResponse.dataExpiracao).toLocaleString(undefined, {
                      year: 'numeric',
                      month: '2-digit',
                      day: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit'
                    })
                  }
                </Typography>
              </Alert>
            </Box>
          )}
        </Paper>
      </Box>
    </LocalizationProvider>
  );
};