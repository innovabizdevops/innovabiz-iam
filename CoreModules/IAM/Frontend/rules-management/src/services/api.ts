/**
 * Serviços de API para gestão de regras dinâmicas
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';

// Configuração da API
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';

/**
 * Cliente HTTP com configurações padrão
 */
export const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  }
});

/**
 * Interceptador de requisições para incluir token de autenticação
 */
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
}, (error) => {
  return Promise.reject(error);
});

/**
 * Interceptador de respostas para tratamento de erros
 */
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Tratamento especial para erro 401 (não autenticado)
    if (error.response && error.response.status === 401) {
      // Redirecionar para página de login
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

/**
 * Métodos genéricos para requisições HTTP
 */
export const api = {
  /**
   * Realiza uma requisição GET
   * 
   * @param url URL da requisição
   * @param config Configurações da requisição
   * @returns Promise com resultado da requisição
   */
  get: <T>(url: string, config?: AxiosRequestConfig) => 
    apiClient.get<T>(url, config).then(response => response.data),

  /**
   * Realiza uma requisição POST
   * 
   * @param url URL da requisição
   * @param data Dados a serem enviados
   * @param config Configurações da requisição
   * @returns Promise com resultado da requisição
   */
  post: <T>(url: string, data?: any, config?: AxiosRequestConfig) => 
    apiClient.post<T>(url, data, config).then(response => response.data),

  /**
   * Realiza uma requisição PUT
   * 
   * @param url URL da requisição
   * @param data Dados a serem enviados
   * @param config Configurações da requisição
   * @returns Promise com resultado da requisição
   */
  put: <T>(url: string, data?: any, config?: AxiosRequestConfig) => 
    apiClient.put<T>(url, data, config).then(response => response.data),

  /**
   * Realiza uma requisição DELETE
   * 
   * @param url URL da requisição
   * @param config Configurações da requisição
   * @returns Promise com resultado da requisição
   */
  delete: <T>(url: string, config?: AxiosRequestConfig) => 
    apiClient.delete<T>(url, config).then(response => response.data),
};