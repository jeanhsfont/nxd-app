import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { CheckCircle, LayoutDashboard } from 'lucide-react';
import api from '../utils/api';

/**
 * Tela de "momento de conexão" após onboarding. Reforça que a fábrica foi conectada
 * e a API Key foi gerada; direciona para o dashboard.
 * Se o usuário já tem telemetria/ativos (acessou /welcome dias depois), redireciona para /.
 */
export default function Welcome() {
  const navigate = useNavigate();
  const [ready, setReady] = useState(false);

  // Já tem dados? Redireciona para dashboard (evita página órfã). Usa api (axios) para 401 cair no interceptor.
  // Request vai para baseURL + path; com baseURL '/' ou origem completa fica exatamente /api/dashboard/data.
  useEffect(() => {
    const token = localStorage.getItem('nxd-token');
    if (!token) {
      setReady(true);
      return;
    }
    let cancelled = false;
    api.get('/api/dashboard/data')
      .then((res) => {
        if (cancelled) return;
        const data = res.data;
        const hasAssets = data?.total_assets > 0 || (data?.assets?.length ?? 0) > 0;
        if (hasAssets) {
          navigate('/', { replace: true });
          return;
        }
        setReady(true);
      })
      .catch(() => {
        if (!cancelled) setReady(true);
      });
    return () => { cancelled = true; };
  }, [navigate]);

  if (!ready) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500 text-sm">Carregando…</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center px-4">
      <div className="max-w-md w-full text-center">
        <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-emerald-100 border border-emerald-200 mb-6">
          <CheckCircle className="w-10 h-10 text-emerald-600" aria-hidden="true" />
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Fábrica conectada com sucesso</h1>
        <p className="text-gray-600 text-sm mb-8">
          Sua chave de API foi gerada. Aguardando primeira telemetria dos seus CLPs. Quando os dados chegarem, eles aparecerão no dashboard.
        </p>
        <Link
          to="/"
          className="inline-flex items-center justify-center gap-2 px-6 py-3 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
        >
          <LayoutDashboard className="w-5 h-5" />
          Ir para Dashboard
        </Link>
      </div>
    </div>
  );
}
