// ==============================================================================
// Nome: BureauRevokeForm.tsx
// Descrição: Formulário para revogação de vínculo com Bureau de Créditos
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
  RadioGroup,
  Radio,
  FormControlLabel,
  FormLabel
} from '@mui/material';
import { useForm, Controller } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { useTranslation } from '../../../hooks/useTranslation';
import { useMultiTenant } from '../../../hooks/useMultiTenant';
import { useQuery } from '@apollo/client';
import { GET_BUREAU_IDENTITY } from '../../../graphql/bureauQueries';

// Esquema de validação com Yup
const revokeVinculoSchema = yup.object().shape({
  motivo: yup.string().required('Campo obrigatório').min(10, 'Explique o motivo com pelo menos 10 caracteres'),
  tipoRevogacao: yup.string().oneOf(['TEMPORARIO', 'PERMANENTE'], 'Tipo de revogação inválido').required('Campo obrigatório'),
  confirmacao: yup.boolean().oneOf([true], 'Você deve confirmar esta ação')
});

interface BureauRevokeFormProps {
  identityId: string;
  onSubmit: (data: {
    identityId: string;
    motivo: string;
    tipoRevogacao: 'TEMPORARIO' | 'PERMANENTE';
  }) => Promise<boolean>;
  loading: boolean;
}

/**
 * Componente de formulário para revogar vínculo com Bureau de Créditos
 * 
 * Este componente permite:
 * - Visualizar informações básicas do vínculo a ser revogado
 * - Selecionar o tipo de revogação (temporário ou permanente)
 * - Fornecer justificativa detalhada para a revogação
 * - Confirmar a ação antes de submeter
 * 
 * Implementa validação avançada e feedback visual de erros
 */
export const BureauRevokeForm: React.FC<BureauRevokeFormProps> = ({
  identityId,
  onSubmit,
  loading
}) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { currentTenant } = useMultiTenant();
  const [formError, setFormError] = useState<string | null>(null);
  const [formSuccess, setFormSuccess] = useState<boolean>(false);

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
    formState: { errors }
  } = useForm({
    resolver: yupResolver(revokeVinculoSchema),
    defaultValues: {
      motivo: '',
      tipoRevogacao: 'TEMPORARIO' as 'TEMPORARIO' | 'PERMANENTE',
      confirmacao: false
    }
  });

  // Função para tratar envio do formulário
  const handleFormSubmit = async (data: any) => {
    try {
      setFormError(null);
      
      // Preparar dados para envio
      const submitData = {
        identityId,
        motivo: data.motivo,
        tipoRevogacao: data.tipoRevogacao
      };

      // Chamar função de submissão passada como prop
      const success = await onSubmit(submitData);
      
      if (success) {
        setFormSuccess(true);
        
        // Resetar formulário após sucesso
        reset();
        
        // Limpar mensagem de sucesso após alguns segundos
        setTimeout(() => {
          setFormSuccess(false);
        }, 5000);
      }
    } catch (error) {
      console.error('Erro ao processar formulário:', error);
      setFormError(t('bureau.erro_processar_revogacao'));
    }
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Paper sx={{ width: '100%', p: 3 }}>
        <Typography variant="h6" gutterBottom color="error">
          {t('bureau.revogar_vinculo_titulo')}
        </Typography>
        <Typography variant="body2" color="textSecondary" paragraph>
          {t('bureau.revogar_vinculo_descricao')}
        </Typography>
        
        {/* Alerta de atenção */}
        <Alert severity="warning" sx={{ mb: 3 }}>
          {t('bureau.revogar_vinculo_alerta')}
        </Alert>

        {/* Mensagem de erro */}
        {formError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {formError}
          </Alert>
        )}

        {/* Mensagem de sucesso */}
        {formSuccess && (
          <Alert severity="success" sx={{ mb: 2 }}>
            {t('bureau.vinculo_revogado_sucesso')}
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
                <Typography variant="subtitle2">{t('bureau.id_vinculo')}</Typography>
                <Typography variant="body2">{identity.id}</Typography>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2">{t('bureau.usuario_vinculo')}</Typography>
                <Typography variant="body2">{identity.usuarioId}</Typography>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2">{t('bureau.tipo_vinculo')}</Typography>
                <Typography variant="body2">{t(`bureau.tipo_${identity.tipoVinculo.toLowerCase()}`)}</Typography>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2">{t('bureau.nivel_acesso')}</Typography>
                <Typography variant="body2">{t(`bureau.nivel_${identity.nivelAcesso.toLowerCase()}`)}</Typography>
              </Grid>
              <Grid item xs={12}>
                <Typography variant="subtitle2">{t('bureau.data_criacao')}</Typography>
                <Typography variant="body2">
                  {new Date(identity.dataCriacao).toLocaleString()}
                </Typography>
              </Grid>
            </Grid>
          </Box>
        ) : null}

        <form onSubmit={handleSubmit(handleFormSubmit)}>
          <Grid container spacing={3}>
            {/* Tipo de revogação */}
            <Grid item xs={12}>
              <Controller
                name="tipoRevogacao"
                control={control}
                render={({ field }) => (
                  <FormControl component="fieldset" error={!!errors.tipoRevogacao}>
                    <FormLabel component="legend">{t('bureau.tipo_revogacao')}</FormLabel>
                    <RadioGroup {...field}>
                      <FormControlLabel 
                        value="TEMPORARIO" 
                        control={<Radio />} 
                        label={t('bureau.revogacao_temporaria')} 
                        disabled={loading}
                      />
                      <FormControlLabel 
                        value="PERMANENTE" 
                        control={<Radio />} 
                        label={t('bureau.revogacao_permanente')} 
                        disabled={loading}
                      />
                    </RadioGroup>
                    {errors.tipoRevogacao && (
                      <FormHelperText error>
                        {errors.tipoRevogacao.message}
                      </FormHelperText>
                    )}
                    <FormHelperText>
                      {field.value === 'TEMPORARIO' 
                        ? t('bureau.revogacao_temporaria_info')
                        : t('bureau.revogacao_permanente_info')}
                    </FormHelperText>
                  </FormControl>
                )}
              />
            </Grid>

            {/* Motivo da revogação */}
            <Grid item xs={12}>
              <Controller
                name="motivo"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    fullWidth
                    label={t('bureau.motivo_revogacao')}
                    placeholder={t('bureau.motivo_revogacao_placeholder')}
                    multiline
                    rows={4}
                    error={!!errors.motivo}
                    helperText={errors.motivo?.message || t('bureau.motivo_revogacao_helper')}
                    disabled={loading}
                  />
                )}
              />
            </Grid>

            {/* Confirmação */}
            <Grid item xs={12}>
              <Controller
                name="confirmacao"
                control={control}
                render={({ field }) => (
                  <FormControl error={!!errors.confirmacao}>
                    <FormControlLabel
                      control={
                        <Radio
                          checked={field.value}
                          onChange={(e) => field.onChange(e.target.checked)}
                          disabled={loading}
                        />
                      }
                      label={
                        <Typography variant="body2" color="error">
                          {t('bureau.confirmo_revogacao')}
                        </Typography>
                      }
                    />
                    {errors.confirmacao && (
                      <FormHelperText error>
                        {errors.confirmacao.message}
                      </FormHelperText>
                    )}
                  </FormControl>
                )}
              />
            </Grid>

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
                  color="error"
                  disabled={loading}
                  startIcon={loading && <CircularProgress size={20} />}
                >
                  {loading ? t('comum.processando') : t('bureau.revogar_vinculo')}
                </Button>
              </Box>
            </Grid>
          </Grid>
        </form>
      </Paper>
    </Box>
  );
};