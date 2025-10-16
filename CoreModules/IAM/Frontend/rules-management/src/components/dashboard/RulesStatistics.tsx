/**
 * Componente de estatísticas de regras para o dashboard
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  Divider,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Typography
} from '@mui/material';
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { useSnackbar } from 'notistack';

import rulesService from '../../services/rulesService';
import { RuleCategory, RuleSeverity } from '../../types/rules';

/**
 * Propriedades do componente
 */
interface RulesStatisticsProps {
  region?: string;
}

/**
 * Dados de estatísticas de regras
 */
interface RuleStatistics {
  matchesByCategory: Record<string, number>;
  matchesBySeverity: Record<string, number>;
  matchesByDay: Array<{ date: string; count: number }>;
  topRules: Array<{ id: string; name: string; matches: number }>;
  totalRules: number;
  activeRules: number;
  totalMatches: number;
  averageScore: number;
}

/**
 * Cores para gráficos
 */
const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#FF6B6B', '#6B8E23', '#9370DB'];
const SEVERITY_COLORS: Record<string, string> = {
  [RuleSeverity.CRITICAL]: '#d32f2f',
  [RuleSeverity.HIGH]: '#f57c00',
  [RuleSeverity.MEDIUM]: '#1976d2',
  [RuleSeverity.LOW]: '#388e3c',
  [RuleSeverity.INFO]: '#757575',
};

/**
 * Componente de estatísticas de regras
 */
const RulesStatistics: React.FC<RulesStatisticsProps> = ({ region }) => {
  // Estados
  const [statistics, setStatistics] = useState<RuleStatistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [timeRange, setTimeRange] = useState<number>(7);
  
  const { enqueueSnackbar } = useSnackbar();
  
  // Buscar estatísticas ao carregar componente e quando filtros mudam
  useEffect(() => {
    fetchStatistics();
  }, [region, timeRange]);
  
  // Função para buscar estatísticas
  const fetchStatistics = async () => {
    try {
      setLoading(true);
      
      // Chamar serviço para obter estatísticas
      const data = await rulesService.getRuleStatistics(region, timeRange);
      
      setStatistics(data);
    } catch (error) {
      console.error('Erro ao buscar estatísticas:', error);
      enqueueSnackbar('Erro ao carregar estatísticas', { variant: 'error' });
    } finally {
      setLoading(false);
    }
  };
  
  // Preparar dados para gráficos
  const prepareChartData = (data: Record<string, number>) => {
    return Object.entries(data).map(([name, value]) => ({
      name,
      value
    }));
  };
  
  // Renderizar indicadores principais
  const renderKeyIndicators = () => {
    if (!statistics) return null;
    
    return (
      <Grid container spacing={3}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Total de Regras
              </Typography>
              <Typography variant="h4">
                {statistics.totalRules}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Regras Ativas
              </Typography>
              <Typography variant="h4">
                {statistics.activeRules}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                {Math.round((statistics.activeRules / statistics.totalRules) * 100)}% do total
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Correspondências
              </Typography>
              <Typography variant="h4">
                {statistics.totalMatches}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Últimos {timeRange} dias
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Score Médio
              </Typography>
              <Typography variant="h4">
                {statistics.averageScore.toFixed(1)}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Escala de 0-100
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h6">
          Estatísticas de Regras{region ? ` - Região: ${region}` : ''}
        </Typography>
        
        <FormControl sx={{ minWidth: 120 }} size="small">
          <InputLabel id="time-range-label">Período</InputLabel>
          <Select
            labelId="time-range-label"
            value={timeRange}
            label="Período"
            onChange={(e) => setTimeRange(Number(e.target.value))}
          >
            <MenuItem value={7}>7 dias</MenuItem>
            <MenuItem value={30}>30 dias</MenuItem>
            <MenuItem value={90}>90 dias</MenuItem>
          </Select>
        </FormControl>
      </Box>
      
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '300px' }}>
          <CircularProgress />
        </Box>
      ) : statistics ? (
        <Box>
          {/* Indicadores principais */}
          {renderKeyIndicators()}
          
          <Grid container spacing={3} sx={{ mt: 2 }}>
            {/* Gráfico de correspondências por dia */}
            <Grid item xs={12} md={8}>
              <Card>
                <CardHeader title="Correspondências por Dia" />
                <Divider />
                <CardContent>
                  <Box sx={{ height: 300 }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart
                        data={statistics.matchesByDay}
                        margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis 
                          dataKey="date" 
                          tickFormatter={(date) => new Date(date).toLocaleDateString('pt-BR', { day: 'numeric', month: 'short' })}
                        />
                        <YAxis />
                        <Tooltip 
                          formatter={(value) => [`${value} correspondências`, 'Contagem']}
                          labelFormatter={(date) => new Date(date).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' })}
                        />
                        <Legend />
                        <Line 
                          type="monotone" 
                          dataKey="count" 
                          name="Correspondências" 
                          stroke="#8884d8" 
                          activeDot={{ r: 8 }} 
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
            
            {/* Gráfico de correspondências por severidade */}
            <Grid item xs={12} md={4}>
              <Card>
                <CardHeader title="Correspondências por Severidade" />
                <Divider />
                <CardContent>
                  <Box sx={{ height: 300 }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={prepareChartData(statistics.matchesBySeverity)}
                          cx="50%"
                          cy="50%"
                          labelLine={false}
                          label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                          outerRadius={80}
                          fill="#8884d8"
                          dataKey="value"
                        >
                          {prepareChartData(statistics.matchesBySeverity).map((entry, index) => (
                            <Cell 
                              key={`cell-${index}`} 
                              fill={SEVERITY_COLORS[entry.name] || COLORS[index % COLORS.length]} 
                            />
                          ))}
                        </Pie>
                        <Tooltip formatter={(value) => [`${value} correspondências`, 'Contagem']} />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
            
            {/* Gráfico de correspondências por categoria */}
            <Grid item xs={12} md={6}>
              <Card>
                <CardHeader title="Correspondências por Categoria" />
                <Divider />
                <CardContent>
                  <Box sx={{ height: 300 }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart
                        data={prepareChartData(statistics.matchesByCategory)}
                        margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="name" />
                        <YAxis />
                        <Tooltip formatter={(value) => [`${value} correspondências`, 'Contagem']} />
                        <Legend />
                        <Bar dataKey="value" name="Correspondências" fill="#82ca9d" />
                      </BarChart>
                    </ResponsiveContainer>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
            
            {/* Top regras mais acionadas */}
            <Grid item xs={12} md={6}>
              <Card>
                <CardHeader title="Top Regras Acionadas" />
                <Divider />
                <CardContent>
                  <Box sx={{ height: 300 }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart
                        data={statistics.topRules}
                        layout="vertical"
                        margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis type="number" />
                        <YAxis 
                          dataKey="name" 
                          type="category"
                          width={150}
                          tick={{ fontSize: 12 }}
                        />
                        <Tooltip formatter={(value) => [`${value} correspondências`, 'Contagem']} />
                        <Legend />
                        <Bar dataKey="matches" name="Correspondências" fill="#8884d8" />
                      </BarChart>
                    </ResponsiveContainer>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </Box>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography variant="body1" color="textSecondary">
            Sem dados estatísticos disponíveis.
          </Typography>
        </Paper>
      )}
    </Box>
  );
};

export default RulesStatistics;