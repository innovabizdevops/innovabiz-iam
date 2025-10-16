/**
 * Componente para teste de regras dinâmicas
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Chip,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Divider,
  Grid,
  Paper,
  TextField,
  Typography
} from '@mui/material';
import {
  Check as CheckIcon,
  Close as CloseIcon,
  PlayArrow as TestIcon
} from '@mui/icons-material';
import { useSnackbar } from 'notistack';

import { Rule, RuleEvaluationResult } from '../../types/rules';
import rulesService from '../../services/rulesService';

/**
 * Propriedades do componente
 */
interface RuleTestProps {
  rule?: Rule;
  open: boolean;
  onClose: () => void;
}

/**
 * Componente para teste de regras
 */
const RuleTest: React.FC<RuleTestProps> = ({
  rule,
  open,
  onClose
}) => {
  // Estados
  const [eventData, setEventData] = useState<string>('{\n  "userId": "user123",\n  "transactionAmount": 1000,\n  "deviceId": "device456",\n  "ipAddress": "192.168.1.1",\n  "timestamp": "2025-08-21T14:30:00Z"\n}');
  const [testResult, setTestResult] = useState<RuleEvaluationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [jsonError, setJsonError] = useState<string | null>(null);
  
  const { enqueueSnackbar } = useSnackbar();
  
  /**
   * Validar JSON
   */
  const validateJson = (json: string): boolean => {
    try {
      JSON.parse(json);
      setJsonError(null);
      return true;
    } catch (error) {
      if (error instanceof Error) {
        setJsonError(error.message);
      } else {
        setJsonError('JSON inválido');
      }
      return false;
    }
  };
  
  /**
   * Testar regra com dados de evento
   */
  const handleTestRule = async () => {
    if (!rule) return;
    
    // Validar JSON
    if (!validateJson(eventData)) {
      return;
    }
    
    try {
      setLoading(true);
      
      // Enviar requisição para testar regra
      const data = await rulesService.testRule(rule.id, JSON.parse(eventData));
      
      setTestResult(data);
      enqueueSnackbar('Regra testada com sucesso', { variant: 'success' });
    } catch (error) {
      console.error('Erro ao testar regra:', error);
      enqueueSnackbar('Erro ao testar regra', { variant: 'error' });
    } finally {
      setLoading(false);
    }
  };
  
  /**
   * Limpar resultado
   */
  const handleClearResult = () => {
    setTestResult(null);
  };
  
  /**
   * Formatar tempo de avaliação
   */
  const formatEvaluationTime = (time?: number): string => {
    if (!time) return 'N/A';
    
    if (time < 1) {
      return `${(time * 1000).toFixed(2)} μs`;
    }
    
    return `${time.toFixed(2)} ms`;
  };

  return (
    <Dialog 
      open={open} 
      onClose={onClose}
      fullWidth
      maxWidth="md"
    >
      <DialogTitle>
        Testar Regra: {rule?.name}
      </DialogTitle>
      
      <DialogContent dividers>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Typography variant="subtitle1" gutterBottom>
              Dados do Evento (JSON)
            </Typography>
            
            <TextField
              fullWidth
              multiline
              rows={8}
              value={eventData}
              onChange={(e) => {
                setEventData(e.target.value);
                if (jsonError) {
                  validateJson(e.target.value);
                }
              }}
              error={!!jsonError}
              helperText={jsonError}
              sx={{ fontFamily: 'monospace' }}
            />
          </Grid>
          
          {testResult && (
            <Grid item xs={12}>
              <Typography variant="subtitle1" gutterBottom>
                Resultado
              </Typography>
              
              <Paper sx={{ p: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6" sx={{ mr: 2 }}>
                    Correspondência:
                  </Typography>
                  
                  {testResult.matched ? (
                    <Chip 
                      icon={<CheckIcon />} 
                      label="Positivo" 
                      color="error" 
                      variant="outlined" 
                    />
                  ) : (
                    <Chip 
                      icon={<CloseIcon />} 
                      label="Negativo" 
                      color="success" 
                      variant="outlined" 
                    />
                  )}
                </Box>
                
                <Divider sx={{ my: 2 }} />
                
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <Typography variant="subtitle2">
                      Score:
                    </Typography>
                    <Typography variant="body1">
                      {testResult.score}
                    </Typography>
                  </Grid>
                  
                  <Grid item xs={12} sm={6}>
                    <Typography variant="subtitle2">
                      Tempo de Avaliação:
                    </Typography>
                    <Typography variant="body1">
                      {formatEvaluationTime(testResult.evaluationTime)}
                    </Typography>
                  </Grid>
                  
                  <Grid item xs={12}>
                    <Typography variant="subtitle2">
                      Ações Acionadas:
                    </Typography>
                    <Box sx={{ mt: 1 }}>
                      {testResult.actions?.length > 0 ? (
                        testResult.actions.map((action, index) => (
                          <Chip 
                            key={index}
                            label={action}
                            size="small"
                            sx={{ mr: 1, mb: 1 }}
                          />
                        ))
                      ) : (
                        <Typography variant="body2" color="textSecondary">
                          Nenhuma ação
                        </Typography>
                      )}
                    </Box>
                  </Grid>
                  
                  {testResult.matchedFields && Object.keys(testResult.matchedFields).length > 0 && (
                    <Grid item xs={12}>
                      <Typography variant="subtitle2" sx={{ mb: 1 }}>
                        Campos Correspondentes:
                      </Typography>
                      <Box sx={{ 
                        p: 1, 
                        backgroundColor: '#f5f5f5', 
                        borderRadius: 1,
                        fontFamily: 'monospace',
                        fontSize: '0.875rem',
                        overflow: 'auto',
                        maxHeight: '150px'
                      }}>
                        <pre>
                          {JSON.stringify(testResult.matchedFields, null, 2)}
                        </pre>
                      </Box>
                    </Grid>
                  )}
                </Grid>
              </Paper>
            </Grid>
          )}
        </Grid>
      </DialogContent>
      
      <DialogActions sx={{ p: 2, justifyContent: 'space-between' }}>
        <Box>
          {testResult && (
            <Button 
              onClick={handleClearResult}
              variant="outlined"
              color="secondary"
            >
              Limpar Resultado
            </Button>
          )}
        </Box>
        
        <Box>
          <Button 
            onClick={onClose}
            variant="outlined"
            sx={{ mr: 1 }}
          >
            Fechar
          </Button>
          
          <Button
            onClick={handleTestRule}
            variant="contained"
            color="primary"
            startIcon={loading ? <CircularProgress size={20} /> : <TestIcon />}
            disabled={loading || !rule}
          >
            {loading ? 'Testando...' : 'Testar Regra'}
          </Button>
        </Box>
      </DialogActions>
    </Dialog>
  );
};

export default RuleTest;