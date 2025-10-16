// ==============================================================================
// Nome: BureauIdentityForm.tsx
// Descrição: Formulário para criação de vínculos com Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React, { useState, useEffect } from 'react';
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
  IconButton,
  Card,
  CardContent,
  CardActions,
  InputAdornment,
  useTheme
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  PersonSearch as PersonSearchIcon
} from '@mui/icons-material';
import { useForm, Controller, useFieldArray } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { useTranslation } from '../../../hooks/useTranslation';
import { useMultiTenant } from '../../../hooks/useMultiTenant';
import { useQuery } from '@apollo/client';
import { SEARCH_USERS_BY_TENANT } from '../graphql/userQueries';
import { TipoVinculo, NivelAcesso } from '../../../types/bureau-credito';

// Esquema de validação com Yup
const identityFormSchema = yup.object().shape({
  usuarioId: yup.string().required('Selecione um usuário'),
  tipoVinculo: yup.string().oneOf(
    ['CONSULTA', 'INTEGRACAO', 'ANALISE'], 
    'Tipo de vínculo inválido'
  ).required('Selecione um tipo de vínculo'),
  nivelAcesso: yup.string().oneOf(
    ['BASICO', 'INTERMEDIARIO', 'COMPLETO'], 
    'Nível de acesso inválido'
  ).required('Selecione um nível de acesso'),
  detalhes: yup.array().of(
    yup.object().shape({
      chave: yup.string().required('Campo obrigatório'),
      valor: yup.string().required('Campo obrigatório')
    })
  )
});

interface BureauIdentityFormProps {
  onSubmit: (data: {
    usuarioId: string;
    tipoVinculo: TipoVinculo;
    nivelAcesso: NivelAcesso;
    detalhes?: Record<string, any>;
  }) => Promise<boolean>;
  loading: boolean;
}

/**
 * Componente de formulário para criar vínculos com Bureau de Créditos
 * 
 * Este componente permite:
 * - Selecionar um usuário do tenant atual
 * - Escolher o tipo de vínculo (Consulta, Integração, Análise)
 * - Definir o nível de acesso (Básico, Intermediário, Completo)
 * - Adicionar campos de detalhes personalizados
 * 
 * Implementa validação avançada, feedback visual de erros e suporte
 * multi-idioma via i18n.
 */
export const BureauIdentityForm: React.FC<BureauIdentityFormProps> = ({
  onSubmit,
  loading
}) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { currentTenant } = useMultiTenant();
  
  const [searchTerm, setSearchTerm] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [formSuccess, setFormSuccess] = useState<boolean>(false);

  // Configuração do React Hook Form com validação Yup
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
    watch
  } = useForm({
    resolver: yupResolver(identityFormSchema),
    defaultValues: {
      usuarioId: '',
      tipoVinculo: 'CONSULTA' as TipoVinculo,
      nivelAcesso: 'BASICO' as NivelAcesso,
      detalhes: [{ chave: '', valor: '' }]
    }
  });

  // Hook para gerenciar o array de campos de detalhes
  const { fields, append, remove } = useFieldArray({
    control,
    name: 'detalhes'
  });

  // Buscar usuários do tenant atual
  const { loading: loadingUsers, error: errorUsers, data: usersData, refetch: refetchUsers } = useQuery(
    SEARCH_USERS_BY_TENANT,
    {
      variables: { 
        tenantId: currentTenant?.id,
        searchTerm: searchTerm || ''
      },
      skip: !currentTenant?.id
    }
  );

  // Atualizar a busca quando o termo de pesquisa mudar
  useEffect(() => {
    const handler = setTimeout(() => {
      if (currentTenant?.id && searchTerm.length > 2) {
        refetchUsers();
      }
    }, 500);

    return () => clearTimeout(handler);
  }, [searchTerm, currentTenant, refetchUsers]);

  // Lista de usuários filtrada
  const users = usersData?.users?.usersByTenant || [];

  // Observar tipo de vínculo para orientação
  const watchTipoVinculo = watch('tipoVinculo');
  const watchNivelAcesso = watch('nivelAcesso');

  // Descrição do tipo de vínculo selecionado
  const getTipoVinculoDescricao = () => {
    switch (watchTipoVinculo) {
      case 'CONSULTA':
        return t('bureau.tipo_consulta_descricao');
      case 'INTEGRACAO':
        return t('bureau.tipo_integracao_descricao');
      case 'ANALISE':
        return t('bureau.tipo_analise_descricao');
      default:
        return '';
    }
  };

  // Descrição do nível de acesso selecionado
  const getNivelAcessoDescricao = () => {
    switch (watchNivelAcesso) {
      case 'BASICO':
        return t('bureau.nivel_basico_descricao');
      case 'INTERMEDIARIO':
        return t('bureau.nivel_intermediario_descricao');
      case 'COMPLETO':
        return t('bureau.nivel_completo_descricao');
      default:
        return '';
    }
  };

  // Função para tratar envio do formulário
  const handleFormSubmit = async (data: any) => {
    try {
      setFormError(null);
      
      // Converter o array de detalhes para objeto
      const detalhesObj: Record<string, any> = {};
      if (data.detalhes && data.detalhes.length > 0) {
        data.detalhes.forEach((detalhe: { chave: string; valor: string }) => {
          if (detalhe.chave && detalhe.valor) {
            detalhesObj[detalhe.chave] = detalhe.valor;
          }
        });
      }

      // Preparar dados para envio
      const submitData = {
        usuarioId: data.usuarioId,
        tipoVinculo: data.tipoVinculo as TipoVinculo,
        nivelAcesso: data.nivelAcesso as NivelAcesso,
        detalhes: Object.keys(detalhesObj).length > 0 ? detalhesObj : undefined
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
      setFormError(t('bureau.erro_criar_vinculo'));
    }
  };

  // Função para adicionar um novo campo de detalhe
  const handleAddDetalhe = () => {
    append({ chave: '', valor: '' });
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Paper sx={{ width: '100%', p: 3 }}>
        <Typography variant="h6" gutterBottom>
          {t('bureau.criar_vinculo_titulo')}
        </Typography>
        <Typography variant="body2" color="textSecondary" paragraph>
          {t('bureau.criar_vinculo_descricao')}
        </Typography>

        {/* Mensagem de erro */}
        {formError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {formError}
          </Alert>
        )}

        {/* Mensagem de erro ao carregar usuários */}
        {errorUsers && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            {t('bureau.erro_carregar_usuarios')}
          </Alert>
        )}

        {/* Mensagem de sucesso */}
        {formSuccess && (
          <Alert severity="success" sx={{ mb: 2 }}>
            {t('bureau.vinculo_criado_sucesso')}
          </Alert>
        )}

        <form onSubmit={handleSubmit(handleFormSubmit)}>
          <Grid container spacing={3}>
            {/* Seleção de usuário */}
            <Grid item xs={12}>
              <Controller
                name="usuarioId"
                control={control}
                render={({ field }) => (
                  <FormControl fullWidth error={!!errors.usuarioId}>
                    <InputLabel id="usuario-select-label">{t('bureau.selecionar_usuario')}</InputLabel>
                    <Select
                      {...field}
                      labelId="usuario-select-label"
                      label={t('bureau.selecionar_usuario')}
                      disabled={loading || !currentTenant}
                      startAdornment={
                        <InputAdornment position="start">
                          <PersonSearchIcon color="action" />
                        </InputAdornment>
                      }
                    >
                      <MenuItem disabled value="">
                        <Box sx={{ p: 1 }}>
                          <TextField
                            size="small"
                            label={t('bureau.buscar_usuario')}
                            variant="outlined"
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            onClick={(e) => e.stopPropagation()}
                            fullWidth
                            InputProps={{
                              startAdornment: (
                                <InputAdornment position="start">
                                  <PersonSearchIcon fontSize="small" />
                                </InputAdornment>
                              ),
                              endAdornment: loadingUsers ? (
                                <CircularProgress size={20} />
                              ) : null
                            }}
                          />
                        </Box>
                      </MenuItem>
                      <Divider />
                      {users.length === 0 ? (
                        <MenuItem disabled>
                          {searchTerm 
                            ? t('bureau.nenhum_usuario_encontrado') 
                            : t('bureau.digite_para_buscar')}
                        </MenuItem>
                      ) : (
                        users.map((user: any) => (
                          <MenuItem key={user.id} value={user.id}>
                            {user.name} ({user.email})
                          </MenuItem>
                        ))
                      )}
                    </Select>
                    {errors.usuarioId && (
                      <FormHelperText error>
                        {errors.usuarioId.message}
                      </FormHelperText>
                    )}
                  </FormControl>
                )}
              />
            </Grid>

            {/* Tipo de vínculo */}
            <Grid item xs={12} md={6}>
              <Controller
                name="tipoVinculo"
                control={control}
                render={({ field }) => (
                  <FormControl fullWidth error={!!errors.tipoVinculo}>
                    <InputLabel id="tipo-vinculo-label">{t('bureau.tipo_vinculo')}</InputLabel>
                    <Select
                      {...field}
                      labelId="tipo-vinculo-label"
                      label={t('bureau.tipo_vinculo')}
                      disabled={loading}
                    >
                      <MenuItem value="CONSULTA">{t('bureau.tipo_consulta')}</MenuItem>
                      <MenuItem value="INTEGRACAO">{t('bureau.tipo_integracao')}</MenuItem>
                      <MenuItem value="ANALISE">{t('bureau.tipo_analise')}</MenuItem>
                    </Select>
                    {errors.tipoVinculo ? (
                      <FormHelperText error>
                        {errors.tipoVinculo.message}
                      </FormHelperText>
                    ) : (
                      <FormHelperText>
                        {getTipoVinculoDescricao()}
                      </FormHelperText>
                    )}
                  </FormControl>
                )}
              />
            </Grid>

            {/* Nível de acesso */}
            <Grid item xs={12} md={6}>
              <Controller
                name="nivelAcesso"
                control={control}
                render={({ field }) => (
                  <FormControl fullWidth error={!!errors.nivelAcesso}>
                    <InputLabel id="nivel-acesso-label">{t('bureau.nivel_acesso')}</InputLabel>
                    <Select
                      {...field}
                      labelId="nivel-acesso-label"
                      label={t('bureau.nivel_acesso')}
                      disabled={loading}
                    >
                      <MenuItem value="BASICO">{t('bureau.nivel_basico')}</MenuItem>
                      <MenuItem value="INTERMEDIARIO">{t('bureau.nivel_intermediario')}</MenuItem>
                      <MenuItem value="COMPLETO">{t('bureau.nivel_completo')}</MenuItem>
                    </Select>
                    {errors.nivelAcesso ? (
                      <FormHelperText error>
                        {errors.nivelAcesso.message}
                      </FormHelperText>
                    ) : (
                      <FormHelperText>
                        {getNivelAcessoDescricao()}
                      </FormHelperText>
                    )}
                  </FormControl>
                )}
              />
            </Grid>

            {/* Campos de detalhes personalizados */}
            <Grid item xs={12}>
              <Typography variant="subtitle2" gutterBottom>
                {t('bureau.detalhes_adicionais')}
              </Typography>
              <Typography variant="body2" color="textSecondary" paragraph>
                {t('bureau.detalhes_adicionais_descricao')}
              </Typography>
              
              {fields.map((field, index) => (
                <Card key={field.id} variant="outlined" sx={{ mb: 2 }}>
                  <CardContent>
                    <Grid container spacing={2}>
                      <Grid item xs={12} md={5}>
                        <Controller
                          name={`detalhes.${index}.chave`}
                          control={control}
                          render={({ field }) => (
                            <TextField
                              {...field}
                              fullWidth
                              label={t('bureau.detalhe_chave')}
                              placeholder={t('bureau.detalhe_chave_placeholder')}
                              error={!!(errors.detalhes && errors.detalhes[index]?.chave)}
                              helperText={
                                errors.detalhes && errors.detalhes[index]?.chave
                                  ? errors.detalhes[index]?.chave?.message
                                  : ''
                              }
                              disabled={loading}
                            />
                          )}
                        />
                      </Grid>
                      <Grid item xs={12} md={7}>
                        <Controller
                          name={`detalhes.${index}.valor`}
                          control={control}
                          render={({ field }) => (
                            <TextField
                              {...field}
                              fullWidth
                              label={t('bureau.detalhe_valor')}
                              placeholder={t('bureau.detalhe_valor_placeholder')}
                              error={!!(errors.detalhes && errors.detalhes[index]?.valor)}
                              helperText={
                                errors.detalhes && errors.detalhes[index]?.valor
                                  ? errors.detalhes[index]?.valor?.message
                                  : ''
                              }
                              disabled={loading}
                            />
                          )}
                        />
                      </Grid>
                    </Grid>
                  </CardContent>
                  <CardActions sx={{ justifyContent: 'flex-end' }}>
                    <Button
                      size="small"
                      color="error"
                      startIcon={<DeleteIcon />}
                      onClick={() => remove(index)}
                      disabled={loading || fields.length <= 1}
                    >
                      {t('comum.remover')}
                    </Button>
                  </CardActions>
                </Card>
              ))}

              <Box sx={{ mt: 2, mb: 2 }}>
                <Button
                  variant="outlined"
                  startIcon={<AddIcon />}
                  onClick={handleAddDetalhe}
                  disabled={loading}
                >
                  {t('bureau.adicionar_detalhe')}
                </Button>
              </Box>
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
                  {loading ? t('comum.processando') : t('bureau.criar_vinculo')}
                </Button>
              </Box>
            </Grid>
          </Grid>
        </form>
      </Paper>
    </Box>
  );
};