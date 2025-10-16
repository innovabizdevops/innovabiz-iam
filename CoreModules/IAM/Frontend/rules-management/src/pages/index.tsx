/**
 * Página principal de gestão de regras dinâmicas
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import React, { useState } from 'react';
import { Box, Container, Tab, Tabs, Typography, Paper } from '@mui/material';
import { SnackbarProvider } from 'notistack';

import RulesList from '../components/rules/RulesList';
import RuleForm from '../components/rules/RuleForm';
import RuleTest from '../components/rules/RuleTest';
import RulesStatistics from '../components/dashboard/RulesStatistics';
import { Rule } from '../types/rules';

/**
 * Propriedade para componentes de painel
 */
interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

/**
 * Componente de painel para abas
 */
function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`rules-tabpanel-${index}`}
      aria-labelledby={`rules-tab-${index}`}
      {...other}
      style={{ width: '100%' }}
    >
      {value === index && (
        <Box sx={{ pt: 2 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

/**
 * Auxiliar para propriedades de abas
 */
function tabProps(index: number) {
  return {
    id: `rules-tab-${index}`,
    'aria-controls': `rules-tabpanel-${index}`,
  };
}

/**
 * Página principal de gestão de regras
 */
export default function RulesManagement() {
  // Estados
  const [selectedTab, setSelectedTab] = useState(0);
  const [selectedRule, setSelectedRule] = useState<Rule | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [isTestingRule, setIsTestingRule] = useState(false);
  
  // Gerenciamento de abas
  const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
    setSelectedTab(newValue);
  };
  
  // Gerenciamento de regras
  const handleRuleSelect = (rule: Rule) => {
    setSelectedRule(rule);
    // Não mudamos para edição automaticamente, apenas selecionamos a regra
  };
  
  const handleRuleEdit = (rule: Rule) => {
    setSelectedRule(rule);
    setIsEditing(true);
    setIsCreating(false);
    setSelectedTab(1); // Mudar para aba de edição
  };
  
  const handleRuleCreate = () => {
    setSelectedRule(null);
    setIsCreating(true);
    setIsEditing(false);
    setSelectedTab(1); // Mudar para aba de criação
  };
  
  const handleRuleTest = (rule: Rule) => {
    setSelectedRule(rule);
    setIsTestingRule(true);
  };
  
  const handleFormCancel = () => {
    setIsEditing(false);
    setIsCreating(false);
    setSelectedTab(0); // Voltar para aba de listagem
  };
  
  const handleFormSave = () => {
    setIsEditing(false);
    setIsCreating(false);
    setSelectedTab(0); // Voltar para aba de listagem
  };
  
  return (
    <SnackbarProvider maxSnack={3}>
      <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
        <Paper sx={{ p: 3, mb: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Gestão de Regras Dinâmicas
          </Typography>
          <Typography variant="body1" paragraph>
            Crie e gerencie regras para detecção de anomalias comportamentais na plataforma INNOVABIZ.
          </Typography>
        </Paper>
        
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs 
            value={selectedTab} 
            onChange={handleTabChange} 
            aria-label="Abas de gestão de regras"
          >
            <Tab label="Regras" {...tabProps(0)} />
            <Tab 
              label={isCreating ? "Nova Regra" : isEditing ? "Editar Regra" : "Detalhes"}
              {...tabProps(1)}
              disabled={!isEditing && !isCreating && !selectedRule}
            />
            <Tab 
              label="Estatísticas" 
              {...tabProps(2)} 
            />
            <Tab 
              label="Conjuntos de Regras" 
              {...tabProps(3)} 
            />
          </Tabs>
        </Box>
        
        <TabPanel value={selectedTab} index={0}>
          <RulesList
            onRuleSelect={handleRuleSelect}
            onRuleEdit={handleRuleEdit}
            onRuleCreate={handleRuleCreate}
            onRuleTest={handleRuleTest}
          />
        </TabPanel>
        
        <TabPanel value={selectedTab} index={1}>
          <RuleForm
            rule={selectedRule || undefined}
            isEdit={isEditing}
            onSave={handleFormSave}
            onCancel={handleFormCancel}
          />
        </TabPanel>
        
        <TabPanel value={selectedTab} index={2}>
          <RulesStatistics />
        </TabPanel>
        
        <TabPanel value={selectedTab} index={3}>
          <Box sx={{ p: 3, textAlign: 'center' }}>
            <Typography variant="h6" color="textSecondary">
              Gestão de Conjuntos de Regras
            </Typography>
            <Typography variant="body1" color="textSecondary">
              Funcionalidade em desenvolvimento...
            </Typography>
          </Box>
        </TabPanel>
        
        {/* Componente de teste de regras */}
        <RuleTest 
          rule={selectedRule || undefined}
          open={isTestingRule}
          onClose={() => setIsTestingRule(false)}
        />
      </Container>
    </SnackbarProvider>
  );
}