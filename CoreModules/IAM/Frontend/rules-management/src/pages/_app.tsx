/**
 * Configuração da aplicação Next.js
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import React from 'react';
import { AppProps } from 'next/app';
import { ThemeProvider, createTheme, ThemeOptions } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { QueryClient, QueryClientProvider } from 'react-query';
import { SnackbarProvider } from 'notistack';

import MainLayout from '../layouts/MainLayout';

// Cria cliente de consulta para React Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutos
    },
  },
});

// Definição do tema da aplicação
const themeOptions: ThemeOptions = {
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
    background: {
      default: '#f5f5f5',
    },
  },
  typography: {
    fontFamily: [
      'Roboto',
      '"Helvetica Neue"',
      'Arial',
      'sans-serif',
    ].join(','),
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        head: {
          fontWeight: 'bold',
        },
      },
    },
  },
};

// Criação do tema
const theme = createTheme(themeOptions);

/**
 * Componente principal da aplicação
 */
function MyApp({ Component, pageProps }: AppProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <SnackbarProvider 
          maxSnack={3} 
          anchorOrigin={{ 
            vertical: 'bottom', 
            horizontal: 'right' 
          }}
          autoHideDuration={5000}
        >
          <MainLayout>
            <Component {...pageProps} />
          </MainLayout>
        </SnackbarProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default MyApp;