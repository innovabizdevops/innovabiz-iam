/**
 * Componente de listagem de regras dinâmicas
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
  Chip,
  IconButton, 
  Button, 
  TextField,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Card,
  CardContent,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Tooltip,
  CircularProgress,
  InputAdornment,
  Grid
} from '@mui/material';
import { DataGrid, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import { 
  Add as AddIcon, 
  Edit as EditIcon, 
  Delete as DeleteIcon,
  PlayArrow as TestIcon,
  Search as SearchIcon,
  FilterList as FilterIcon
} from '@mui/icons-material';
import { useSnackbar } from 'notistack';

import { Rule, RuleFilter, RuleSeverity, RuleCategory } from '../../types/rules';
import rulesService from '../../services/rulesService';

/**
 * Propriedades do componente
 */
interface RulesListProps {
  region?: string;
  onRuleSelect?: (rule: Rule) => void;
  onRuleCreate?: () => void;
  onRuleEdit?: (rule: Rule) => void;
  onRuleTest?: (rule: Rule) => void;
}

/**
 * Componente de listagem de regras
 */
export const RulesList: React.FC<RulesListProps> = ({
  region,
  onRuleSelect,
  onRuleCreate,
  onRuleEdit,
  onRuleTest
}) => {
  // Estados
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<RuleFilter>({
    region: region || undefined,
    searchTerm: '',
    category: undefined,
    severity: undefined,
    enabled: undefined
  });
  const [selectedRule, setSelectedRule] = useState<Rule | null>(null);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showFilters, setShowFilters] = useState(false);
  
  const { enqueueSnackbar } = useSnackbar();
  
  // Buscar regras ao carregar componente e quando filtros mudam
  useEffect(() => {
    fetchRules();
  }, [region, filters.enabled, filters.category, filters.severity]);
  
  // Função para buscar regras
  const fetchRules = async () => {
    try {
      setLoading(true);
      const data = await rulesService.getRules({
        region: filters.region,
        tags: filters.tags
      });
      
      // Aplicar filtros do cliente (já que a API não suporta todos)
      let filteredData = [...data];
      
      if (filters.searchTerm) {
        const searchTerm = filters.searchTerm.toLowerCase();
        filteredData = filteredData.filter(rule => 
          rule.name.toLowerCase().includes(searchTerm) || 
          (rule.description && rule.description.toLowerCase().includes(searchTerm))
        );
      }
      
      if (filters.category !== undefined) {
        filteredData = filteredData.filter(rule => rule.category === filters.category);
      }
      
      if (filters.severity !== undefined) {
        filteredData = filteredData.filter(rule => rule.severity === filters.severity);
      }
      
      if (filters.enabled !== undefined) {
        filteredData = filteredData.filter(rule => rule.enabled === filters.enabled);
      }
      
      setRules(filteredData);
    } catch (error) {
      console.error('Erro ao buscar regras:', error);
      enqueueSnackbar('Erro ao carregar regras', { variant: 'error' });
    } finally {
      setLoading(false);
    }
  };
  
  // Função para excluir regra
  const handleDeleteRule = async () => {
    if (!selectedRule) return;
    
    try {
      await rulesService.deleteRule(selectedRule.id);
      
      enqueueSnackbar('Regra excluída com sucesso', { variant: 'success' });
      setShowDeleteDialog(false);
      
      // Atualizar lista de regras
      await fetchRules();
    } catch (error) {
      console.error('Erro ao excluir regra:', error);
      enqueueSnackbar('Erro ao excluir regra', { variant: 'error' });
    }
  };
  
  // Função para confirmar exclusão de regra
  const confirmDeleteRule = (rule: Rule) => {
    setSelectedRule(rule);
    setShowDeleteDialog(true);
  };
  
  // Função para limpar filtros
  const clearFilters = () => {
    setFilters({
      region: region || undefined,
      searchTerm: '',
      category: undefined,
      severity: undefined,
      enabled: undefined
    });
  };
  
  // Colunas da grid
  const columns: GridColDef[] = [
    { 
      field: 'name', 
      headerName: 'Nome', 
      flex: 1, 
      minWidth: 200 
    },
    { 
      field: 'description', 
      headerName: 'Descrição', 
      flex: 2,
      minWidth: 250,
      renderCell: (params: GridRenderCellParams) => (
        <Tooltip title={params.value || ''}>
          <span>{params.value || ''}</span>
        </Tooltip>
      )
    },
    { 
      field: 'severity', 
      headerName: 'Severidade', 
      width: 150,
      renderCell: (params: GridRenderCellParams) => {
        const severity = params.value as RuleSeverity;
        const color = 
          severity === RuleSeverity.CRITICAL ? 'error' :
          severity === RuleSeverity.HIGH ? 'warning' :
          severity === RuleSeverity.MEDIUM ? 'primary' :
          severity === RuleSeverity.LOW ? 'info' : 'default';
        
        return <Chip label={severity} color={color} size="small" />;
      }
    },
    { 
      field: 'category', 
      headerName: 'Categoria', 
      width: 160,
      renderCell: (params: GridRenderCellParams) => (
        <Chip label={params.value} size="small" variant="outlined" />
      )
    },
    { 
      field: 'region', 
      headerName: 'Região', 
      width: 120,
      renderCell: (params: GridRenderCellParams) => (
        <span>{params.value || 'Global'}</span>
      )
    },
    { 
      field: 'enabled', 
      headerName: 'Status', 
      width: 130,
      renderCell: (params: GridRenderCellParams) => (
        <Chip 
          label={params.value ? 'Ativo' : 'Inativo'} 
          color={params.value ? 'success' : 'default'} 
          size="small"
          variant="outlined"
        />
      )
    },
    { 
      field: 'actions', 
      headerName: 'Ações', 
      width: 160,
      sortable: false,
      filterable: false,
      renderCell: (params: GridRenderCellParams) => {
        const rule = params.row as Rule;
        
        return (
          <Box>
            <Tooltip title="Testar regra">
              <IconButton 
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  onRuleTest?.(rule);
                }}
              >
                <TestIcon fontSize="small" />
              </IconButton>
            </Tooltip>
            
            <Tooltip title="Editar regra">
              <IconButton 
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  onRuleEdit?.(rule);
                }}
              >
                <EditIcon fontSize="small" />
              </IconButton>
            </Tooltip>
            
            <Tooltip title="Excluir regra">
              <IconButton 
                size="small"
                color="error"
                onClick={(e) => {
                  e.stopPropagation();
                  confirmDeleteRule(rule);
                }}
              >
                <DeleteIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          </Box>
        );
      }
    },
  ];

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6" component="h2">
            Regras de Detecção de Anomalias
          </Typography>
          
          <Box>
            <Button
              startIcon={<FilterIcon />}
              onClick={() => setShowFilters(!showFilters)}
              size="small"
              sx={{ mr: 1 }}
            >
              Filtros
            </Button>
            
            <Button
              variant="contained"
              color="primary"
              startIcon={<AddIcon />}
              onClick={() => onRuleCreate?.()}
            >
              Nova Regra
            </Button>
          </Box>
        </Box>
        
        {showFilters && (
          <Paper sx={{ p: 2, mb: 2 }}>
            <Grid container spacing={2} alignItems="center">
              <Grid item xs={12} sm={4} md={3}>
                <TextField
                  label="Pesquisar"
                  fullWidth
                  size="small"
                  value={filters.searchTerm || ''}
                  onChange={(e) => setFilters({...filters, searchTerm: e.target.value})}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchIcon fontSize="small" />
                      </InputAdornment>
                    ),
                  }}
                />
              </Grid>
              
              <Grid item xs={12} sm={4} md={3}>
                <FormControl fullWidth size="small">
                  <InputLabel>Categoria</InputLabel>
                  <Select
                    label="Categoria"
                    value={filters.category || ''}
                    onChange={(e) => setFilters({...filters, category: e.target.value as RuleCategory || undefined})}
                  >
                    <MenuItem value="">Todas</MenuItem>
                    {Object.values(RuleCategory).map((category) => (
                      <MenuItem key={category} value={category}>
                        {category}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              
              <Grid item xs={12} sm={4} md={2}>
                <FormControl fullWidth size="small">
                  <InputLabel>Severidade</InputLabel>
                  <Select
                    label="Severidade"
                    value={filters.severity || ''}
                    onChange={(e) => setFilters({...filters, severity: e.target.value as RuleSeverity || undefined})}
                  >
                    <MenuItem value="">Todas</MenuItem>
                    {Object.values(RuleSeverity).map((severity) => (
                      <MenuItem key={severity} value={severity}>
                        {severity}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              
              <Grid item xs={12} sm={4} md={2}>
                <FormControl fullWidth size="small">
                  <InputLabel>Status</InputLabel>
                  <Select
                    label="Status"
                    value={filters.enabled === undefined ? '' : (filters.enabled ? 'true' : 'false')}
                    onChange={(e) => {
                      const value = e.target.value;
                      setFilters({
                        ...filters, 
                        enabled: value === '' ? undefined : value === 'true'
                      });
                    }}
                  >
                    <MenuItem value="">Todos</MenuItem>
                    <MenuItem value="true">Ativo</MenuItem>
                    <MenuItem value="false">Inativo</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              
              <Grid item xs={12} sm={4} md={2}>
                <Button 
                  variant="outlined" 
                  onClick={clearFilters}
                  fullWidth
                >
                  Limpar Filtros
                </Button>
              </Grid>
            </Grid>
          </Paper>
        )}
        
        <Box sx={{ height: 500, width: '100%' }}>
          <DataGrid
            rows={rules}
            columns={columns}
            loading={loading}
            paginationModel={{
              pageSize: 10,
              page: 0,
            }}
            pageSizeOptions={[5, 10, 25]}
            checkboxSelection={false}
            disableRowSelectionOnClick
            onRowClick={(params) => {
              onRuleSelect?.(params.row as Rule);
            }}
          />
        </Box>
      </CardContent>
      
      {/* Diálogo de confirmação de exclusão */}
      <Dialog open={showDeleteDialog} onClose={() => setShowDeleteDialog(false)}>
        <DialogTitle>Confirmar Exclusão</DialogTitle>
        <DialogContent>
          <Typography>
            Tem certeza que deseja excluir a regra &quot;{selectedRule?.name}&quot;? Esta ação não pode ser desfeita.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowDeleteDialog(false)}>Cancelar</Button>
          <Button onClick={handleDeleteRule} color="error" variant="contained">
            Excluir
          </Button>
        </DialogActions>
      </Dialog>
    </Card>
  );
};

export default RulesList;