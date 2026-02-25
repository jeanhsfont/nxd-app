import React, { useState } from 'react';
import { AlertTriangle, Copy, CheckCircle } from 'lucide-react';
import api from '../utils/api';

export default function RegenerateKeyModal({ onClose }) {
  const [step, setStep] = useState('confirm'); // 'confirm' | 'generated'
  const [newKey, setNewKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState('');

  const handleRegenerate = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await api.post('/api/factory/regenerate-api-key');
      setNewKey(response.data.apiKey || response.data.api_key);
      setStep('generated');
    } catch (err) {
      setError(err.response?.data?.error || 'Erro ao gerar nova chave. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(newKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (step === 'generated') {
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
        <div className="bg-white rounded-xl shadow-2xl p-8 w-full max-w-lg mx-4">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-green-100 flex items-center justify-center">
              <CheckCircle className="w-6 h-6 text-green-600" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-gray-900">Nova API Key Gerada!</h2>
              <p className="text-sm text-gray-500">Copie e guarde em local seguro</p>
            </div>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6 flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
            <div className="text-sm text-yellow-800">
              <p className="font-semibold mb-1">Importante!</p>
              <p>Sua chave anterior foi invalidada. Esta nova chave será exibida apenas uma vez.</p>
            </div>
          </div>

          <div className="bg-gray-900 rounded-lg p-4 mb-6">
            <p className="text-xs text-gray-400 mb-2">Nova API Key:</p>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-sm text-white font-mono break-all">{newKey}</code>
              <button
                onClick={handleCopy}
                className="flex-shrink-0 bg-gray-700 hover:bg-gray-600 text-white px-3 py-2 rounded-lg flex items-center gap-2 transition-colors"
              >
                {copied ? (
                  <>
                    <CheckCircle className="w-4 h-4" />
                    <span className="text-xs">Copiado!</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    <span className="text-xs">Copiar</span>
                  </>
                )}
              </button>
            </div>
          </div>

          <button
            onClick={onClose}
            className="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-3 px-4 rounded-lg transition-colors"
          >
            Fechar
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="bg-white rounded-xl shadow-2xl p-8 w-full max-w-lg mx-4">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-12 h-12 rounded-xl bg-red-100 flex items-center justify-center">
            <AlertTriangle className="w-6 h-6 text-red-600" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">Gerar Nova API Key?</h2>
            <p className="text-sm text-gray-500">Ação irreversível</p>
          </div>
        </div>

        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800 font-medium mb-2">Atenção:</p>
          <ul className="text-sm text-red-700 space-y-1 list-disc list-inside">
            <li>Sua API Key atual será <strong>invalidada permanentemente</strong></li>
            <li>Todos os dispositivos conectados perderão acesso</li>
            <li>Você precisará reconfigurar todos os dispositivos com a nova chave</li>
          </ul>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold py-3 px-4 rounded-lg transition-colors"
            disabled={loading}
          >
            Cancelar
          </button>
          <button
            onClick={handleRegenerate}
            disabled={loading}
            className="flex-1 bg-red-600 hover:bg-red-700 text-white font-semibold py-3 px-4 rounded-lg transition-colors disabled:opacity-50"
          >
            {loading ? 'Gerando...' : 'Sim, gerar nova chave'}
          </button>
        </div>
      </div>
    </div>
  );
}
