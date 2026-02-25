import React, { useState } from 'react';
import { Copy, CheckCircle, AlertTriangle } from 'lucide-react';

export default function ApiKeyModal({ apiKey, onClose }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="nxd-card w-full max-w-lg mx-4">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-12 h-12 rounded-lg bg-green/10 flex items-center justify-center">
            <CheckCircle className="w-6 h-6 text-green" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">API Key Gerada!</h2>
            <p className="text-sm text-gray-500">Copie e guarde em local seguro</p>
          </div>
        </div>

        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6 flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
          <div className="text-sm text-yellow-800">
            <p className="font-semibold mb-1">Importante!</p>
            <p>Esta chave será exibida apenas uma vez. Copie agora e guarde em local seguro.</p>
          </div>
        </div>

        <div className="bg-gray-900 rounded-lg p-4 mb-6">
          <p className="text-xs text-gray-400 mb-2">Sua API Key:</p>
          <div className="flex items-center gap-2">
            <code className="flex-1 text-sm text-white font-mono break-all">{apiKey}</code>
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

        <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-6">
          <p className="text-xs text-gray-600 mb-2">Como usar:</p>
          <ol className="text-xs text-gray-600 space-y-1 list-decimal list-inside">
            <li>Configure esta chave no seu DX Simulator ou dispositivo</li>
            <li>Inicie o envio de telemetria</li>
            <li>Acompanhe os dados no Dashboard</li>
          </ol>
        </div>

        <button
          onClick={onClose}
          className="nxd-btn nxd-btn-primary w-full justify-center"
        >
          Entendi, continuar
        </button>
      </div>
    </div>
  );
}
