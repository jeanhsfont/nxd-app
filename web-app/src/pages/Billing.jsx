import React, { useEffect, useState } from 'react';
import { CreditCard, CheckCircle, ExternalLink } from 'lucide-react';
import api from '../utils/api';

const PLAN_LABELS = { free: 'Grátis', pro: 'NXD Pro', enterprise: 'NXD Enterprise' };

export default function Billing() {
  const [plan, setPlan] = useState(null);
  const [nextBilling, setNextBilling] = useState('');
  const [checkoutUrl, setCheckoutUrl] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    document.title = 'Cobrança | NXD';
    return () => { document.title = 'NXD'; };
  }, []);

  useEffect(() => {
    api.get('/api/billing/plan')
      .then((res) => {
        setPlan(res.data.plan || 'free');
        setNextBilling(res.data.next_billing || '');
        setCheckoutUrl(res.data.checkout_url || '');
      })
      .catch(() => setError('Não foi possível carregar o plano.'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="p-8 max-w-4xl mx-auto">
        <div className="animate-pulse h-8 bg-gray-200 rounded w-48 mb-6" />
        <div className="h-40 bg-gray-100 rounded-xl" />
      </div>
    );
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-6">
        <CreditCard className="w-8 h-8 text-indigo-600" />
        <h1 className="text-2xl font-bold text-gray-900">Cobrança</h1>
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
          {error}
        </div>
      )}

      <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden mb-8">
        <div className="p-6 border-b border-gray-100">
          <h2 className="text-lg font-semibold text-gray-800">Plano atual</h2>
          <p className="text-gray-500 text-sm mt-1">
            {PLAN_LABELS[plan] || plan} — 1 fábrica, até 50 ativos
          </p>
          <div className="mt-4 flex items-center gap-2">
            <CheckCircle className="w-5 h-5 text-emerald-500" />
            <span className="text-sm text-emerald-600 font-medium">
              {plan === 'free' ? 'Plano gratuito' : 'Assinatura ativa'}
            </span>
          </div>
        </div>
        <div className="p-6 grid grid-cols-2 gap-4 text-sm">
          {nextBilling && (
            <div>
              <span className="text-gray-500">Próxima cobrança</span>
              <p className="font-medium text-gray-900">{nextBilling}</p>
            </div>
          )}
          {checkoutUrl && (
            <div className="col-span-2">
              <a
                href={checkoutUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
              >
                <ExternalLink className="w-4 h-4" />
                Pagar / Renovar assinatura
              </a>
            </div>
          )}
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded-xl shadow-sm p-6">
        <h2 className="text-lg font-semibold text-gray-800 mb-2">Planos</h2>
        <ul className="text-sm text-gray-600 space-y-2">
          <li><strong>Grátis:</strong> 1 fábrica, até 10 ativos.</li>
          <li><strong>NXD Pro:</strong> 1 fábrica, até 50 ativos, suporte por e-mail.</li>
          <li><strong>NXD Enterprise:</strong> múltiplas fábricas, ativos ilimitados, suporte prioritário.</li>
        </ul>
        {!checkoutUrl && (
          <p className="mt-4 text-gray-600 text-sm">
            Upgrade de plano estará disponível em breve. Entre em contato com o suporte para mais informações.
          </p>
        )}
      </div>
    </div>
  );
}
