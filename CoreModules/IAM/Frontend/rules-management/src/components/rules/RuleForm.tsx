/**
 * Componente de formulário para criação/edição de regras dinâmicas
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  Card,
  CardContent,
  CardHeader,
  TextField, 
  Button, 
  FormControl,
  FormHelperText,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  Divider,
  Grid,
  IconButton,
  Chip,
  Autocomplete,
  Tooltip
} from '@mui/material';
import { 
  Add as AddIcon, 
  Delete as DeleteIcon,
  Save as SaveIcon,
  ArrowBack as BackIcon
} from '@mui/icons-material';
import { useSnackbar } from 'notistack';
import { Formik, Form, FieldArray } from 'formik';
import * as Yup from 'yup';

import { 
  Rule, RuleCondition, RuleGroup, 
  RuleOperator, RuleLogicalOperator, RuleAction,
  RuleSeverity, RuleCategory, RuleValueType 
} from '../../types/rules';
import rulesService from '../../services/rulesService';

/**
 * Propriedades do componente
 */
interface RuleFormProps {
  rule?: Rule;
  isEdit?: boolean;
  region?: string;
  onSave?: (rule: Rule) => void;
  onCancel?: () => void;
}

/**
 * Validação do formulário
 */
const validationSchema = Yup.object().shape({
  name: Yup.string().required('Nome é obrigatório'),
  description: Yup.string(),
  enabled: Yup.boolean(),
  severity: Yup.string().required('Severidade é obrigatória'),
  category: Yup.string().required('Categoria é obrigatória'),
  region: Yup.string(),
  tags: Yup.array().of(Yup.string()),
  actions: Yup.array().of(Yup.string()).min(1, 'Pelo menos uma ação é obrigatória'),
  condition: Yup.object().when('group', {
    is: undefined,
    then: Yup.object({
      field: Yup.string().required('Campo é obrigatório'),
      operator: Yup.string().required('Operador é obrigatório'),
      value: Yup.mixed().required('Valor é obrigatório')
    }).required('Condição simples ou grupo é obrigatório'),
    otherwise: Yup.object().nullable()
  }),
  group: Yup.object().when('condition', {
    is: undefined,
    then: Yup.object({
      operator: Yup.string().required('Operador lógico é obrigatório'),
      conditions: Yup.array().of(
        Yup.object({
          field: Yup.string().required('Campo é obrigatório'),
          operator: Yup.string().required('Operador é obrigatório'),
          value: Yup.mixed().required('Valor é obrigatório')
        })
      ).min(1, 'Pelo menos uma condição é obrigatória')
    }).required('Condição simples ou grupo é obrigatório'),
    otherwise: Yup.object().nullable()
  }),
  score: Yup.number().min(0).max(100)
});

/**
 * Valor inicial para nova regra
 */
const initialRule: Partial<Rule> = {
  name: '',
  description: '',
  enabled: true,
  severity: RuleSeverity.MEDIUM,
  category: RuleCategory.BEHAVIORAL,
  region: '',
  tags: [],
  condition: {
    field: '',
    operator: RuleOperator.EQUALS,
    value: ''
  },
  actions: [RuleAction.LOG],
  score: 50
};

/**
 * Componente de formulário de regra
 */
export const RuleForm: React.FC<RuleFormProps> = ({
  rule,
  isEdit = false,
  region = '',
  onSave,
  onCancel
}) => {
  const { enqueueSnackbar } = useSnackbar();
  const [useConditionGroup, setUseConditionGroup] = useState<boolean>(
    !!rule?.group || false
  );
  
  // Preparar valores iniciais
  const initialValues: Partial<Rule> = isEdit && rule 
    ? { ...rule } 
    : { ...initialRule, region };
  
  // Salvar regra
  const handleSubmit = async (values: Partial<Rule>) => {
    try {
      let savedRule: Rule;
      
      // Verificar se é uma condição simples ou grupo
      if (!useConditionGroup) {
        // Usar condição simples
        delete values.group;
      } else {
        // Usar grupo de condições
        delete values.condition;
      }
      
      if (isEdit && rule) {
        // Atualizar regra existente
        savedRule = await rulesService.updateRule(rule.id, values);
        enqueueSnackbar('Regra atualizada com sucesso', { variant: 'success' });
      } else {
        // Criar nova regra
        savedRule = await rulesService.createRule(values);
        enqueueSnackbar('Regra criada com sucesso', { variant: 'success' });
      }
      
      onSave?.(savedRule);
    } catch (error) {
      console.error('Erro ao salvar regra:', error);
      enqueueSnackbar('Erro ao salvar regra', { variant: 'error' });
    }
  };
  
  return (
    <Card>
      <CardHeader 
        title={isEdit ? "Editar Regra" : "Nova Regra"} 
        action={
          <Button 
            startIcon={<BackIcon />} 
            onClick={onCancel}
          >
            Voltar
          </Button>
        }
      />
      <CardContent>
        <Formik
          initialValues={initialValues as Rule}
          validationSchema={validationSchema}
          onSubmit={handleSubmit}
        >
          {({ values, errors, touched, handleChange, handleBlur, setFieldValue, isSubmitting }) => (
            <Form>
              <Grid container spacing={3}>
                {/* Informações básicas */}
                <Grid item xs={12}>
                  <Typography variant="subtitle1" gutterBottom>
                    Informações Básicas
                  </Typography>
                  <Paper sx={{ p: 2 }}>
                    <Grid container spacing={2}>
                      <Grid item xs={12} sm={8}>
                        <TextField
                          name="name"
                          label="Nome"
                          fullWidth
                          value={values.name}
                          onChange={handleChange}
                          onBlur={handleBlur}
                          error={touched.name && !!errors.name}
                          helperText={touched.name && errors.name}
                          required
                        />
                      </Grid>
                      
                      <Grid item xs={12} sm={4}>
                        <FormControlLabel
                          control={
                            <Switch
                              name="enabled"
                              checked={values.enabled}
                              onChange={handleChange}
                              color="primary"
                            />
                          }
                          label="Ativa"
                        />
                      </Grid>
                      
                      <Grid item xs={12}>
                        <TextField
                          name="description"
                          label="Descrição"
                          fullWidth
                          multiline
                          rows={2}
                          value={values.description || ''}
                          onChange={handleChange}
                          onBlur={handleBlur}
                        />
                      </Grid>
                      
                      <Grid item xs={12} sm={6} md={3}>
                        <FormControl fullWidth error={touched.severity && !!errors.severity}>
                          <InputLabel id="severity-label">Severidade</InputLabel>
                          <Select
                            labelId="severity-label"
                            name="severity"
                            value={values.severity}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            required
                          >
                            {Object.values(RuleSeverity).map((severity) => (
                              <MenuItem key={severity} value={severity}>
                                {severity}
                              </MenuItem>
                            ))}
                          </Select>
                          {touched.severity && errors.severity && (
                            <FormHelperText>{errors.severity}</FormHelperText>
                          )}
                        </FormControl>
                      </Grid>
                      
                      <Grid item xs={12} sm={6} md={3}>
                        <FormControl fullWidth error={touched.category && !!errors.category}>
                          <InputLabel id="category-label">Categoria</InputLabel>
                          <Select
                            labelId="category-label"
                            name="category"
                            value={values.category}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            required
                          >
                            {Object.values(RuleCategory).map((category) => (
                              <MenuItem key={category} value={category}>
                                {category}
                              </MenuItem>
                            ))}
                          </Select>
                          {touched.category && errors.category && (
                            <FormHelperText>{errors.category}</FormHelperText>
                          )}
                        </FormControl>
                      </Grid>
                      
                      <Grid item xs={12} sm={6} md={3}>
                        <TextField
                          name="region"
                          label="Região"
                          fullWidth
                          value={values.region || ''}
                          onChange={handleChange}
                          onBlur={handleBlur}
                          placeholder="Global"
                          helperText="Deixe em branco para global"
                        />
                      </Grid>
                      
                      <Grid item xs={12} sm={6} md={3}>
                        <TextField
                          name="score"
                          label="Score"
                          fullWidth
                          type="number"
                          inputProps={{ min: 0, max: 100 }}
                          value={values.score || 0}
                          onChange={handleChange}
                          onBlur={handleBlur}
                          error={touched.score && !!errors.score}
                          helperText={touched.score && errors.score}
                        />
                      </Grid>
                      
                      <Grid item xs={12}>
                        <Autocomplete
                          multiple
                          freeSolo
                          options={[]}
                          value={values.tags || []}
                          onChange={(_, newValue) => {
                            setFieldValue('tags', newValue);
                          }}
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
                              label="Tags"
                              placeholder="Adicionar tag"
                              helperText="Pressione Enter para adicionar"
                            />
                          )}
                        />
                      </Grid>
                    </Grid>
                  </Paper>
                </Grid>
                
                {/* Condições */}
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                    <Typography variant="subtitle1">
                      Condições
                    </Typography>
                    
                    <FormControlLabel
                      control={
                        <Switch
                          checked={useConditionGroup}
                          onChange={(e) => {
                            setUseConditionGroup(e.target.checked);
                            
                            if (e.target.checked) {
                              // Mudar para grupo de condições
                              setFieldValue('group', {
                                operator: RuleLogicalOperator.AND,
                                conditions: values.condition ? [values.condition] : []
                              });
                              setFieldValue('condition', undefined);
                            } else {
                              // Mudar para condição simples
                              const firstCondition = values.group?.conditions?.[0];
                              setFieldValue('condition', firstCondition || {
                                field: '',
                                operator: RuleOperator.EQUALS,
                                value: ''
                              });
                              setFieldValue('group', undefined);
                            }
                          }}
                        />
                      }
                      label="Usar grupo de condições"
                    />
                  </Box>
                  
                  <Paper sx={{ p: 2 }}>
                    {!useConditionGroup ? (
                      // Condição simples
                      <Grid container spacing={2}>
                        <Grid item xs={12} sm={4}>
                          <TextField
                            name="condition.field"
                            label="Campo"
                            fullWidth
                            value={values.condition?.field || ''}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            error={touched.condition?.field && !!errors.condition?.field}
                            helperText={touched.condition?.field && errors.condition?.field}
                            required
                          />
                        </Grid>
                        
                        <Grid item xs={12} sm={4}>
                          <FormControl fullWidth>
                            <InputLabel>Operador</InputLabel>
                            <Select
                              name="condition.operator"
                              value={values.condition?.operator || ''}
                              onChange={handleChange}
                              onBlur={handleBlur}
                              required
                            >
                              {Object.values(RuleOperator).map((op) => (
                                <MenuItem key={op} value={op}>
                                  {op}
                                </MenuItem>
                              ))}
                            </Select>
                          </FormControl>
                        </Grid>
                        
                        <Grid item xs={12} sm={4}>
                          <TextField
                            name="condition.value"
                            label="Valor"
                            fullWidth
                            value={values.condition?.value || ''}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            error={touched.condition?.value && !!errors.condition?.value}
                            helperText={touched.condition?.value && errors.condition?.value}
                            required
                          />
                        </Grid>
                      </Grid>
                    ) : (
                      // Grupo de condições
                      <Box>
                        <FormControl fullWidth sx={{ mb: 2 }}>
                          <InputLabel>Operador Lógico</InputLabel>
                          <Select
                            name="group.operator"
                            value={values.group?.operator || ''}
                            onChange={handleChange}
                            onBlur={handleBlur}
                            required
                          >
                            {Object.values(RuleLogicalOperator).map((op) => (
                              <MenuItem key={op} value={op}>
                                {op}
                              </MenuItem>
                            ))}
                          </Select>
                        </FormControl>
                        
                        <FieldArray name="group.conditions">
                          {({ remove, push }) => (
                            <Box>
                              {values.group?.conditions?.map((_, index) => (
                                <Paper key={index} sx={{ p: 2, mb: 2 }}>
                                  <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <Typography variant="subtitle2">
                                      Condição {index + 1}
                                    </Typography>
                                    
                                    <IconButton
                                      size="small"
                                      color="error"
                                      onClick={() => remove(index)}
                                    >
                                      <DeleteIcon />
                                    </IconButton>
                                  </Box>
                                  
                                  <Grid container spacing={2}>
                                    <Grid item xs={12} sm={4}>
                                      <TextField
                                        name={`group.conditions[${index}].field`}
                                        label="Campo"
                                        fullWidth
                                        value={values.group?.conditions?.[index]?.field || ''}
                                        onChange={handleChange}
                                        onBlur={handleBlur}
                                        required
                                      />
                                    </Grid>
                                    
                                    <Grid item xs={12} sm={4}>
                                      <FormControl fullWidth>
                                        <InputLabel>Operador</InputLabel>
                                        <Select
                                          name={`group.conditions[${index}].operator`}
                                          value={values.group?.conditions?.[index]?.operator || ''}
                                          onChange={handleChange}
                                          onBlur={handleBlur}
                                          required
                                        >
                                          {Object.values(RuleOperator).map((op) => (
                                            <MenuItem key={op} value={op}>
                                              {op}
                                            </MenuItem>
                                          ))}
                                        </Select>
                                      </FormControl>
                                    </Grid>
                                    
                                    <Grid item xs={12} sm={4}>
                                      <TextField
                                        name={`group.conditions[${index}].value`}
                                        label="Valor"
                                        fullWidth
                                        value={values.group?.conditions?.[index]?.value || ''}
                                        onChange={handleChange}
                                        onBlur={handleBlur}
                                        required
                                      />
                                    </Grid>
                                  </Grid>
                                </Paper>
                              ))}
                              
                              <Button
                                startIcon={<AddIcon />}
                                onClick={() => push({
                                  field: '',
                                  operator: RuleOperator.EQUALS,
                                  value: ''
                                })}
                              >
                                Adicionar Condição
                              </Button>
                            </Box>
                          )}
                        </FieldArray>
                      </Box>
                    )}
                  </Paper>
                </Grid>
                
                {/* Ações */}
                <Grid item xs={12}>
                  <Typography variant="subtitle1" gutterBottom>
                    Ações
                  </Typography>
                  <Paper sx={{ p: 2 }}>
                    <FormControl 
                      fullWidth 
                      error={touched.actions && !!errors.actions}
                      required
                    >
                      <InputLabel id="actions-label">Ações</InputLabel>
                      <Select
                        labelId="actions-label"
                        name="actions"
                        multiple
                        value={values.actions || []}
                        onChange={handleChange}
                        onBlur={handleBlur}
                        renderValue={(selected) => (
                          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                            {(selected as string[]).map((value) => (
                              <Chip key={value} label={value} size="small" />
                            ))}
                          </Box>
                        )}
                      >
                        {Object.values(RuleAction).map((action) => (
                          <MenuItem key={action} value={action}>
                            {action}
                          </MenuItem>
                        ))}
                      </Select>
                      {touched.actions && errors.actions && (
                        <FormHelperText>{errors.actions as string}</FormHelperText>
                      )}
                    </FormControl>
                  </Paper>
                </Grid>
                
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
                    <Button 
                      variant="outlined" 
                      onClick={onCancel}
                      sx={{ mr: 2 }}
                    >
                      Cancelar
                    </Button>
                    
                    <Button 
                      type="submit"
                      variant="contained" 
                      color="primary"
                      startIcon={<SaveIcon />}
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? 'Salvando...' : isEdit ? 'Atualizar' : 'Salvar'}
                    </Button>
                  </Box>
                </Grid>
              </Grid>
            </Form>
          )}
        </Formik>
      </CardContent>
    </Card>
  );
};

export default RuleForm;