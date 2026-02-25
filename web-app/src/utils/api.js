import axios from 'axios';

// Backend Go (HubSystem1.0): use REACT_APP_BACKEND_URL ou deixe vazio para mesmo host
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || '';

const api = axios.create({
  baseURL: BACKEND_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor para adicionar token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('nxd-token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor para tratar erros de autenticação
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expirado ou inválido
      localStorage.removeItem('nxd-token');
      
      // Marca no sessionStorage para mostrar toast
      if (typeof sessionStorage !== 'undefined') {
        sessionStorage.setItem('nxd_session_expired_toast', 'true');
      }
      
      // Redireciona para login
      window.location.href = '/login?reason=session_expired';
    }
    return Promise.reject(error);
  }
);

export default api;
