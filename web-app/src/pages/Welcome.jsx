import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { CheckCircle, LayoutDashboard, ArrowRight, Zap } from 'lucide-react';
import api from '../utils/api';

export default function Welcome() {
  const navigate = useNavigate();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('nxd-token');
    if (!token) {
      setReady(true);
      return;
    }
    api.get('/api/dashboard/data')
      .then((res) => {
        const hasAssets = res.data?.total_assets > 0;
        if (hasAssets) navigate('/', { replace: true });
        else setReady(true);
      })
      .catch(() => setReady(true));
  }, [navigate]);

  if (!ready) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-2xl w-full text-center fade-in">
        <div className="inline-flex items-center justify-center w-20 h-20 bg-navy rounded-xl mb-6">
          <CheckCircle className="w-10 h-10 text-white" />
        </div>

        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Fábrica conectada
        </h1>
        <p className="text-gray-600 text-lg mb-2">
          Sua chave de API foi gerada com sucesso.
        </p>
        <p className="text-navy font-semibold mb-8">
          Em poucos minutos você pode ver seu primeiro ROI no painel de indicadores financeiros.
        </p>

        <div className="nxd-card max-w-xl mx-auto mb-8 text-left">
          <h3 className="text-gray-900 font-semibold mb-4 text-lg">Próximos Passos</h3>
          <div className="space-y-3">
            {[
              'Vá em Configurações e copie sua API Key',
              'Configure no DX Simulator',
              'Inicie o envio de telemetria',
              'Acompanhe dados em tempo real no Dashboard'
            ].map((step, i) => (
              <div key={i} className="flex items-center gap-3 text-gray-700">
                <div className="w-7 h-7 bg-navy rounded-full flex items-center justify-center text-white font-bold text-sm flex-shrink-0">
                  {i + 1}
                </div>
                <span>{step}</span>
              </div>
            ))}
          </div>
        </div>

        <Link to="/">
          <button className="nxd-btn nxd-btn-primary">
            <LayoutDashboard className="w-5 h-5" />
            Ir para Dashboard
            <ArrowRight className="w-5 h-5" />
          </button>
        </Link>
      </div>
    </div>
  );
}
