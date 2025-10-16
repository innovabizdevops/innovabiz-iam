// ==============================================================================
// Nome: BureauAuthorizationForm.tsx
// Descrição: Formulário para criação de autorizações de consulta ao Bureau de Créditos
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
  InputLabel,
  Select,
  MenuItem,
  FormHelperText,
  CircularProgress,
  Alert,
  Divider,
  useTheme,
  Chip,
  Stack
} from '@mui/material';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { ptBR, enUS } from 'date-fns/locale';
import { useForm, Controller } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { useTranslation } from '../../../hooks/useTranslation';
import { useMultiTenant } from '../../../hooks/useMultiTenant';
import { useLanguage } from '../../../hooks/useLanguage';
import { addMinutes, addHours, addDays } from 'date-fns';

// Tipos
import { TipoConsulta } from '../../../types/bureau-credito';

// Esquema de validação com Yup
const createAutorizacaoSchema = yup.object().shape({
  vinculoId: yup.string().required('Campo obrigatório'),
  tipoConsulta: yup.string().oneOf(['SIMPLES', 'COMPLETA', 'SCORE', 'ANALITICA'], 'Tipo inválido').required('Campo obrigatório'),
  finalidade: yup.string().required('Campo obrigatório').min(10, 'Descreva com pelo menos 10 caracteres'),
  justificativa: yup.string().required('Campo obrigatório').min(20, 'Justificativa deve ter pelo menos 20 caracteres'),
  dataValidade: yup.date().required('Campo obrigatório').min(new Date(), 'Data deve ser no futuro')
});

interface BureauAuthorizationFormProps {
  vinculoId: string;
  onSubmit: (data: {
    vinculoId: string;
    tipoConsulta: TipoConsulta;
    finalidade: string;
    justificativa: string;
    dataValidade: Date;
    tags?: string[];
  }) => Promise<boolean>;
  loading: boolean;
}

/**
 * Componente de formulário para criar nova autorização de consulta ao Bureau de Créditos
 * 
 * Este componente permite:
 * - Selecionar tipo de consulta (simples, completa, score, analítica)
 * - Definir finalidade da consulta
 * - Fornecer justificativa detalhada
 * - Configurar data de validade
 * - Adicionar tags para classificação e busca
 * 
 * Implementa validação avançada e feedback visual de erros
 */
export const BureauAuthorizationForm: React.FC<BureauAuthorizationFormProps> = ({ 
  vinculoId, 
  onSubmit, 
  loading 
}) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { currentTenant } = useMultiTenant();
  const { language } = useLanguage();
  const [formError, setFormError] = useState<string | null>(null);
  const [formSuccess, setFormSuccess] = useState<boolean>(false);
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState<string>('');

  // Seleção de locale para os componentes de data
  const dateLocale = language === 'pt' ? ptBR : enUS;

  // Configuração do React Hook Form com validação Yup
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
    setValue
  } = useForm({
    resolver: yupResolver(createAutorizacaoSchema),
    defaultValues: {
      vinculoId: vinculoId,
      tipoConsulta: 'SIMPLES' as TipoConsulta,
      finalidade: '',
      justificativa: '',
      dataValidade: addDays(new Date(), 7) // Padrão: 7 dias
    }
  });

  // Observar valor atual do tipo de consulta
  const tipoConsultaValue = watch('tipoConsulta');

  // Função para adicionar tag
  const handleAddTag = () => {
    if (newTag.trim() && !tags.includes(newTag.trim())) {
      setTags([...tags, newTag.trim()]);
      setNewTag('');
    }
  };

  // Função para remover tag
  const handleDeleteTag = (tagToDelete: string) => {
    setTags(tags.filter((tag) => tag !== tagToDelete));
  };

  // Funções para ajustar data de validade rapidamente
  const handleSetValidity = (hours: number) => {
    setValue('dataValidade', addHours(new Date(), hours));
  };

  // Função para tratar envio do formulário
  const handleFormSubmit = async (data: any) => {
    try {
      setFormError(null);
      
      // Preparar dados para envio com tags
      const submitData = {
        ...data,
        tags
      };

      // Chamar função de submissão passada como prop
      const success = await onSubmit(submitData);
      
      if (success) {
        setFormSuccess(true);
        
        // Resetar formulário após sucesso
        reset();
        setTags([]);
        
        // Limpar mensagem de sucesso após alguns segundos
        setTimeout(() => {
          setFormSuccess(false);
        }, 5000);
      }
    } catch (error) {
      console.error('Erro ao processar formulário:', error);
      setFormError(t('bureau.erro_processar_autorizacao'));
    }
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Paper sx={{ width: '100%', p: 3 }}>
        <Typography variant="h6" gutterBottom>
          {t('bureau.nova_autorizacao_titulo')}
        </Typography>
        <Typography variant="body2" color="textSecondary" paragraph>
          {t('bureau.nova_autorizacao_descricao')}
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

        <form onSubmit={handleSubmit(handleFormSubmit)}>
          <Grid container spacing={3}>
            {/* Tipo de consulta */}
            <Grid item xs={12}>
              <Controller
                name="tipoConsulta"
                control={control}
                render={({ field }) => (
                  <FormControl fullWidth error={!!errors.tipoConsulta}>
                    <InputLabel id="tipo-consulta-label">
                      {t('bureau.tipo_consulta')}
                    </InputLabel>
                    <Select
                      {...field}
                      labelId="tipo-consulta-label"
                      label={t('bureau.tipo_consulta')}
                      disabled={loading}
                    >
                      <MenuItem value="SIMPLES">{t('bureau.consulta_simples')}</MenuItem>
                      <MenuItem value="COMPLETA">{t('bureau.consulta_completa')}</MenuItem>
                      <MenuItem value="SCORE">{t('bureau.consulta_score')}</MenuItem>
                      <MenuItem value="ANALITICA">{t('bureau.consulta_analitica')}</MenuItem>
                    </Select>
                    {errors.tipoConsulta && (
                      <FormHelperText error>
                        {errors.tipoConsulta.message}
                      </FormHelperText>
                    )}
                    <FormHelperText>
                      {tipoConsultaValue === 'SIMPLES' && t('bureau.descricao_consulta_simples')}
                      {tipoConsultaValue === 'COMPLETA' && t('bureau.descricao_consulta_completa')}
                      {tipoConsultaValue === 'SCORE' && t('bureau.descricao_consulta_score')}
                      {tipoConsultaValue === 'ANALITICA' && t('bureau.descricao_consulta_analitica')}
                    </FormHelperText>
                  </FormControl>
                )}
              />
            </Grid>

            {/* Finalidade */}
            <Grid item xs={12}>
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

            {/* Justificativa */}
            <Grid item xs={12}>
              <Controller
                name="justificativa"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    fullWidth
                    label={t('bureau.justificativa')}
                    placeholder={t('bureau.justificativa_placeholder')}
                    multiline
                    rows={4}
                    error={!!errors.justificativa}
                    helperText={errors.justificativa?.message || t('bureau.justificativa_helper')}
                    disabled={loading}
                  />
                )}
              />
            </Grid>

            {/* Data de validade */}
            <Grid item xs={12}>
              <LocalizationProvider dateAdapter={AdapterDateFns} adapterLocale={dateLocale}>
                <Controller
                  name="dataValidade"
                  control={control}
                  render={({ field }) => (
                    <DateTimePicker
                      label={t('bureau.data_validade')}
                      value={field.value}
                      onChange={(date) => field.onChange(date)}
                      slotProps={{
                        textField: {
                          fullWidth: true,
                          error: !!errors.dataValidade,
                          helperText: errors.dataValidade?.message,
                        },
                      }}
                      disablePast
                      disabled={loading}
                    />
                  )}
                />
              </LocalizationProvider>

              {/* Botões de atalho para definir validade */}
              <Box sx={{ mt: 1, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                <Button 
                  size="small" 
                  variant="outlined" 
                  onClick={() => handleSetValidity(1)}
                  disabled={loading}
                >
                  1h
                </Button>
                <Button 
                  size="small" 
                  variant="outlined" 
                  onClick={() => handleSetValidity(6)}
                  disabled={loading}
                >
                  6h
                </Button>
                <Button 
                  size="small" 
                  variant="outlined" 
                  onClick={() => handleSetValidity(24)}
                  disabled={loading}
                >
                  1d
                </Button>
                <Button 
                  size="small" 
                  variant="outlined" 
                  onClick={() => handleSetValidity(24 * 7)}
                  disabled={loading}
                >
                  7d
                </Button>
                <Button 
                  size="small" 
                  variant="outlined" 
                  onClick={() => handleSetValidity(24 * 30)}
                  disabled={loading}
                >
                  30d
                </Button>
              </Box>
            </Grid>

            {/* Tags */}
            <Grid item xs={12}>
              <Typography variant="subtitle2" gutterBottom>
                {t('bureau.tags')}
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <TextField
                  value={newTag}
                  onChange={(e) => setNewTag(e.target.value)}
                  label={t('bureau.nova_tag')}
                  variant="outlined"
                  size="small"
                  disabled={loading}
                  sx={{ mr: 1 }}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddTag();
                    }
                  }}
                />
                <Button
                  variant="contained"
                  size="small"
                  onClick={handleAddTag}
                  disabled={loading || !newTag.trim()}
                >
                  {t('comum.adicionar')}
                </Button>
              </Box>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                {tags.map((tag) => (
                  <Chip
                    key={tag}
                    label={tag}
                    onDelete={() => handleDeleteTag(tag)}
                    color="primary"
                    variant="outlined"
                    size="small"
                    sx={{ mb: 1 }}
                  />
                ))}
                {tags.length === 0 && (
                  <Typography variant="body2" color="textSecondary">
                    {t('bureau.nenhuma_tag')}
                  </Typography>
                )}
              </Stack>
              <FormHelperText>
                {t('bureau.tags_helper')}
              </FormHelperText>
            </Grid>

            <Grid item xs={12}>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 2 }}>
                <Button
                  variant="outlined"
                  onClick={() => {
                    reset();
                    setTags([]);
                  }}
                  disabled={loading}
                >
                  {t('comum.limpar')}
                </Button>
                <Button
                  type="submit"
                  variant="contained"
                  color="primary"
                  disabled={loading}
                  startIcon={loading && <CircularProgress size={20} />}
                >
                  {loading ? t('comum.processando') : t('comum.criar_autorizacao')}
                </Button>
              </Box>
            </Grid>
          </Grid>
        </form>
      </Paper>
    </Box>
  );
};