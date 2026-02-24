import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/',
});

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

// Guard: evita duplo redirect quando várias requisições recebem 401 em paralelo
let handling401 = false;

// Tratamento global de token expirado (401) — enterprise
// Ordem obrigatória: 1) hadToken (antes de removeItem) 2) early return 3) removeItem 4) toast 5) redirect
// Só faz fluxo "sessão expirada" se havia token (evita false positive em /terms, /support, etc.)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status !== 401) return Promise.reject(error);

    const hadToken = !!localStorage.getItem('nxd-token');
    if (!hadToken) return Promise.reject(error);

    const currentPath = window.location.pathname;
    const isAuthPage = currentPath === '/login' || currentPath === '/register';
    if (isAuthPage) return Promise.reject(error);

    if (handling401) return Promise.reject(error);
    handling401 = true;

    localStorage.removeItem('nxd-token');
    try {
      sessionStorage.setItem('nxd_session_expired_toast', '1');
    } catch (_) {}
    window.location.replace('/login?reason=session_expired');
    return Promise.reject(error);
  }
);

export default api;
