import React, { useEffect, useState } from 'react';
import { CreditCard, CheckCircle, Zap, Crown, Sparkles } from 'lucide-react';
import api from '../utils/api';
import toast from 'react-hot-toast';

const PLANS = [
  {
    id: 'free',
    name: 'Free',
    price: 'R$ 0',
    period: '/mês',
    icon: Zap,
    features: ['1 fábrica', 'Até 10 ativos', 'Dashboard básico'],
  },
  {
    id: 'pro',
    name: 'Pro',
    price: 'R$ 199',
    period: '/mês',
    icon: Crown,
    popular: true,
    features: ['1 fábrica', 'Até 50 ativos', 'IA Intelligence', 'Relatórios', 'Suporte prioritário'],
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: 'Customizado',
    period: '',
    icon: Sparkles,
    features: ['Múltiplas fábricas', 'Ativos ilimitados', 'Suporte 24/7', 'SLA garantido'],
  },
];

export default function Billing() {
  const [billing, setBilling] = useState({ plan: 'free', status: 'active', next_billing: '', max_assets: 5, checkout_url: '' });
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState(null);

  useEffect(() => {
    api.get('/api/billing/plan')
      .then((res) => setBilling({
        plan: res.data.plan || 'free',
        status: res.data.status || 'active',
        next_billing: res.data.next_billing || '',
        max_assets: res.data.max_assets ?? 5,
        checkout_url: res.data.checkout_url || '',
      }))
      .catch(() => toast.error('Erro ao carregar plano'))
      .finally(() => setLoading(false));
  }, []);

  const handleAssinar = async (planId) => {
    if (planId === 'free' || planId === billing.plan) return;
    setCheckoutLoading(planId);
    try {
      const successUrl = window.location.origin + '/billing?success=1';
      const cancelUrl = window.location.origin + '/billing';
      const res = await api.post('/api/billing/create-checkout-session', {
        plan: planId,
        success_url: successUrl,
        cancel_url: cancelUrl,
      });
      if (res.data?.url) {
        window.location.href = res.data.url;
      } else {
        toast.error(res.data?.error || 'Checkout não configurado');
      }
    } catch (e) {
      toast.error(e.response?.data?.error || 'Erro ao iniciar checkout');
    } finally {
      setCheckoutLoading(null);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto p-6">
        <div className="page-header">
          <div className="page-header-icon">
            <CreditCard className="w-6 h-6" />
          </div>
          <div>
            <h1 className="page-title">Planos e Cobrança</h1>
            <p className="page-subtitle">
              Plano atual: <span className="text-navy font-semibold">{billing.plan}</span>
              {billing.status && billing.status !== 'active' && (
                <span className="ml-2 text-amber-600">Status: {billing.status}</span>
              )}
              {' · '}Limite: {billing.max_assets === 99999 ? 'ilimitado' : billing.max_assets} ativos
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {PLANS.map((plan) => (
            <div
              key={plan.id}
              className={`nxd-card relative fade-in ${
                billing.plan === plan.id ? 'ring-2 ring-navy' : ''
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-navy text-white text-xs font-bold px-4 py-1 rounded-full">
                  POPULAR
                </div>
              )}
              {billing.plan === plan.id && (
                <div className="absolute -top-3 right-4 bg-green text-white text-xs font-bold px-4 py-1 rounded-full flex items-center gap-1">
                  <CheckCircle className="w-3 h-3" /> ATUAL
                </div>
              )}

              <div className="w-12 h-12 bg-navy rounded-lg flex items-center justify-center mb-4">
                <plan.icon className="w-6 h-6 text-white" />
              </div>

              <h3 className="text-xl font-bold text-gray-900 mb-2">{plan.name}</h3>
              <div className="mb-6">
                <span className="text-3xl font-black text-navy">{plan.price}</span>
                <span className="text-gray-500">{plan.period}</span>
              </div>

              <ul className="space-y-2 mb-8">
                {plan.features.map((f, i) => (
                  <li key={i} className="flex items-start gap-2 text-gray-700 text-sm">
                    <CheckCircle className="w-4 h-4 text-green flex-shrink-0 mt-0.5" />
                    <span>{f}</span>
                  </li>
                ))}
              </ul>

              <button
                disabled={billing.plan === plan.id || checkoutLoading !== null}
                onClick={() => handleAssinar(plan.id)}
                className={`nxd-btn w-full justify-center ${
                  billing.plan === plan.id
                    ? 'nxd-btn-primary opacity-50 cursor-not-allowed'
                    : 'nxd-btn-primary'
                }`}
              >
                {billing.plan === plan.id
                  ? 'Plano Atual'
                  : plan.id === 'enterprise'
                    ? 'Falar com Vendas'
                    : checkoutLoading === plan.id ? 'Redirecionando...' : 'Assinar'}
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
