// ==============================================================================
// Nome: BureauIdentitiesList.tsx
// Descrição: Componente de listagem de vínculos com Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Chip,
  IconButton,
  Tooltip,
  TextField,
  MenuItem,
  Grid,
  InputAdornment,
  Button,
  CircularProgress,
  Alert,
  useTheme
} from '@mui/material';
import {
  Visibility as VisibilityIcon,
  Refresh as RefreshIcon,
  Search as SearchIcon,
  FilterList as FilterListIcon,
  SortByAlpha as SortIcon
} from '@mui/icons-material';
import { format } from 'date-fns';
import { ptBR } from 'date-fns/locale';
import { useTranslation } from '../../../hooks/useTranslation';
import { usePermissions } from '../../../hooks/usePermissions';
import { 
  BureauIdentity, 
  TipoVinculo, 
  BureauVinculoStatus 
} from '../../../types/bureau-credito';

// Props do componente
interface BureauIdentitiesListProps {
  loading: boolean;
  error: any;
  identities: BureauIdentity[];
  onSelect: (identityId: string) => void;
  filter: {
    status: BureauVinculoStatus | '';
    tipoVinculo: TipoVinculo | '';
  };
  onFilterChange: (filter: {
    status: BureauVinculoStatus | '';
    tipoVinculo: TipoVinculo | '';
  }) => void;
  onRefresh: () => void;
}

/**
 * Componente para listagem de vínculos com Bureau de Créditos
 * 
 * Este componente exibe:
 * - Lista paginada de vínculos do Bureau de Créditos
 * - Filtros por status e tipo de vínculo
 * - Busca por texto nos detalhes do vínculo
 * - Exibição de status visual com chips coloridos
 * - Botões de ação para visualização e outras operações
 * 
 * Implementa paginação no lado do cliente e ordenação
 */
export const BureauIdentitiesList: React.FC<BureauIdentitiesListProps> = ({
  loading,
  error,
  identities,
  onSelect,
  filter,
  onFilterChange,
  onRefresh
}) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { hasPermission } = usePermissions();
  
  // Estado local para paginação e busca
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [sortField, setSortField] = useState<'dataCriacao' | 'status'>('dataCriacao');

  // Permissões
  const canViewDetails = hasPermission('bureau_credito:view_details');
  
  // Funções de manipulação para paginação
  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  // Função para alternar a ordenação
  const handleSort = (field: 'dataCriacao' | 'status') => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('asc');
    }
  };

  // Função para aplicar filtros, ordenação e busca
  const filteredAndSortedIdentities = React.useMemo(() => {
    // Aplicar filtros
    let filtered = [...identities];
    
    if (filter.status) {
      filtered = filtered.filter(identity => identity.status === filter.status);
    }
    
    if (filter.tipoVinculo) {
      filtered = filtered.filter(identity => identity.tipoVinculo === filter.tipoVinculo);
    }
    
    // Aplicar busca
    if (searchTerm) {
      const searchLower = searchTerm.toLowerCase();
      filtered = filtered.filter(identity => 
        identity.id.toLowerCase().includes(searchLower) ||
        identity.usuarioNome?.toLowerCase().includes(searchLower) ||
        identity.detalhes?.some(d => 
          d.chave.toLowerCase().includes(searchLower) || 
          d.valor.toLowerCase().includes(searchLower)
        )
      );
    }
    
    // Aplicar ordenação
    filtered.sort((a, b) => {
      if (sortField === 'dataCriacao') {
        const dateA = new Date(a.dataCriacao).getTime();
        const dateB = new Date(b.dataCriacao).getTime();
        return sortOrder === 'asc' ? dateA - dateB : dateB - dateA;
      } else {
        // Ordenar por status
        const statusOrder = {
          ATIVO: 0,
          PENDENTE: 1,
          SUSPENSO: 2,
          REVOGADO: 3
        };
        
        const statusA = statusOrder[a.status as keyof typeof statusOrder];
        const statusB = statusOrder[b.status as keyof typeof statusOrder];
        
        return sortOrder === 'asc' 
          ? statusA - statusB 
          : statusB - statusA;
      }
    });
    
    return filtered;
  }, [identities, filter, searchTerm, sortField, sortOrder]);

  // Determinar cor e texto do chip de status
  const getStatusChipProps = (status: BureauVinculoStatus) => {
    switch(status) {
      case 'ATIVO':
        return { 
          color: 'success' as const, 
          label: t('bureau.status_ativo'),
          icon: '✓'
        };
      case 'PENDENTE':
        return { 
          color: 'warning' as const, 
          label: t('bureau.status_pendente'),
          icon: '⌛'
        };
      case 'SUSPENSO':
        return { 
          color: 'info' as const, 
          label: t('bureau.status_suspenso'),
          icon: '⏸'
        };
      case 'REVOGADO':
        return { 
          color: 'error' as const, 
          label: t('bureau.status_revogado'),
          icon: '✕'
        };
      default:
        return { 
          color: 'default' as const, 
          label: status,
          icon: '?'
        };
    }
  };
  
  // Páginas de resultados para paginação
  const paginatedIdentities = filteredAndSortedIdentities.slice(
    page * rowsPerPage, 
    page * rowsPerPage + rowsPerPage
  );

  // Formatar data
  const formatDate = (dateString: string) => {
    try {
      return format(new Date(dateString), 'PPpp', { locale: ptBR });
    } catch (error) {
      return dateString;
    }
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Paper sx={{ width: '100%', p: 2 }}>
        <Typography variant="h6" gutterBottom>
          {t('bureau.vinculos_titulo')}
        </Typography>

        {/* Filtros e busca */}
        <Grid container spacing={2} sx={{ mb: 3 }}>
          {/* Busca */}
          <Grid item xs={12} sm={6} md={4}>
            <TextField
              fullWidth
              size="small"
              label={t('comum.buscar')}
              variant="outlined"
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value);
                setPage(0); // Reset para primeira página
              }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon />
                  </InputAdornment>
                ),
              }}
              placeholder={t('bureau.buscar_placeholder')}
            />
          </Grid>

          {/* Filtro por Status */}
          <Grid item xs={12} sm={6} md={3}>
            <TextField
              select
              fullWidth
              size="small"
              label={t('comum.status')}
              variant="outlined"
              value={filter.status}
              onChange={(e) => {
                onFilterChange({
                  ...filter,
                  status: e.target.value as BureauVinculoStatus | ''
                });
                setPage(0); // Reset para primeira página
              }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <FilterListIcon />
                  </InputAdornment>
                ),
              }}
            >
              <MenuItem value="">{t('comum.todos')}</MenuItem>
              <MenuItem value="ATIVO">{t('bureau.status_ativo')}</MenuItem>
              <MenuItem value="PENDENTE">{t('bureau.status_pendente')}</MenuItem>
              <MenuItem value="SUSPENSO">{t('bureau.status_suspenso')}</MenuItem>
              <MenuItem value="REVOGADO">{t('bureau.status_revogado')}</MenuItem>
            </TextField>
          </Grid>

          {/* Filtro por Tipo */}
          <Grid item xs={12} sm={6} md={3}>
            <TextField
              select
              fullWidth
              size="small"
              label={t('bureau.tipo_vinculo')}
              variant="outlined"
              value={filter.tipoVinculo}
              onChange={(e) => {
                onFilterChange({
                  ...filter,
                  tipoVinculo: e.target.value as TipoVinculo | ''
                });
                setPage(0); // Reset para primeira página
              }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <FilterListIcon />
                  </InputAdornment>
                ),
              }}
            >
              <MenuItem value="">{t('comum.todos')}</MenuItem>
              <MenuItem value="CONSULTA">{t('bureau.tipo_consulta')}</MenuItem>
              <MenuItem value="INTEGRACAO">{t('bureau.tipo_integracao')}</MenuItem>
              <MenuItem value="ANALISE">{t('bureau.tipo_analise')}</MenuItem>
            </TextField>
          </Grid>

          {/* Botão de Atualizar */}
          <Grid item xs={12} sm={6} md={2} sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Button
              variant="outlined"
              color="primary"
              onClick={onRefresh}
              disabled={loading}
              startIcon={loading ? <CircularProgress size={20} /> : <RefreshIcon />}
              fullWidth
            >
              {t('comum.atualizar')}
            </Button>
          </Grid>
        </Grid>

        {/* Mensagem de erro */}
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {t('bureau.erro_carregar_vinculos')}
          </Alert>
        )}

        {/* Tabela de resultados */}
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>{t('bureau.id')}</TableCell>
                <TableCell>{t('bureau.usuario')}</TableCell>
                <TableCell>{t('bureau.tipo')}</TableCell>
                <TableCell>{t('bureau.nivel_acesso')}</TableCell>
                <TableCell 
                  onClick={() => handleSort('status')}
                  sx={{ cursor: 'pointer' }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    {t('comum.status')}
                    {sortField === 'status' && (
                      <SortIcon 
                        fontSize="small" 
                        sx={{ ml: 0.5, transform: sortOrder === 'desc' ? 'rotate(180deg)' : 'none' }}
                      />
                    )}
                  </Box>
                </TableCell>
                <TableCell 
                  onClick={() => handleSort('dataCriacao')}
                  sx={{ cursor: 'pointer' }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    {t('bureau.data_criacao')}
                    {sortField === 'dataCriacao' && (
                      <SortIcon 
                        fontSize="small" 
                        sx={{ ml: 0.5, transform: sortOrder === 'desc' ? 'rotate(180deg)' : 'none' }}
                      />
                    )}
                  </Box>
                </TableCell>
                <TableCell align="center">{t('comum.acoes')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={7} align="center" sx={{ py: 3 }}>
                    <CircularProgress size={30} />
                  </TableCell>
                </TableRow>
              ) : paginatedIdentities.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} align="center" sx={{ py: 3 }}>
                    <Typography variant="body1" color="textSecondary">
                      {searchTerm || filter.status || filter.tipoVinculo
                        ? t('bureau.nenhum_resultado')
                        : t('bureau.nenhum_vinculo')}
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                paginatedIdentities.map((identity) => {
                  const statusChipProps = getStatusChipProps(identity.status);
                  
                  return (
                    <TableRow
                      key={identity.id}
                      hover
                      onClick={() => canViewDetails && onSelect(identity.id)}
                      sx={{ cursor: canViewDetails ? 'pointer' : 'default' }}
                    >
                      <TableCell>
                        <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                          {identity.id.substring(0, 8)}...
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {identity.usuarioNome || t('comum.nao_disponivel')}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {t(`bureau.tipo_${identity.tipoVinculo.toLowerCase()}`)}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {t(`bureau.nivel_${identity.nivelAcesso.toLowerCase()}`)}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip 
                          size="small"
                          label={statusChipProps.label}
                          color={statusChipProps.color}
                          icon={<span>{statusChipProps.icon}</span>}
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {formatDate(identity.dataCriacao)}
                        </Typography>
                      </TableCell>
                      <TableCell align="center">
                        <Box>
                          <Tooltip title={t('comum.visualizar')}>
                            <span>
                              <IconButton
                                size="small"
                                color="primary"
                                disabled={!canViewDetails}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  onSelect(identity.id);
                                }}
                              >
                                <VisibilityIcon fontSize="small" />
                              </IconButton>
                            </span>
                          </Tooltip>
                        </Box>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </TableContainer>

        {/* Paginação */}
        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={filteredAndSortedIdentities.length}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
          labelRowsPerPage={t('comum.itens_por_pagina')}
          labelDisplayedRows={({ from, to, count }) => 
            `${from}-${to} ${t('comum.de')} ${count}`
          }
        />
      </Paper>
    </Box>
  );
};