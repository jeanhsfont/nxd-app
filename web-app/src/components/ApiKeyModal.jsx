import React, { useState, useRef, useEffect } from 'react';

const CopyIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
  </svg>
);

const InfoIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
);

function getFocusables(container) {
  if (!container) return [];
  const selector = 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])';
  return Array.from(container.querySelectorAll(selector)).filter(
    (el) => !el.disabled && el.offsetParent !== null
  );
}

export default function ApiKeyModal({ apiKey, onClose }) {
  const [hasAcknowledged, setHasAcknowledged] = useState(false);
  const [copySuccess, setCopySuccess] = useState('');
  const containerRef = useRef(null);
  const previousActiveRef = useRef(null);

  const handleCopy = () => {
    navigator.clipboard.writeText(apiKey);
    setCopySuccess('Chave copiada para a área de transferência!');
    setTimeout(() => setCopySuccess(''), 2000);
  };

  // Focus trap e foco inicial
  useEffect(() => {
    previousActiveRef.current = document.activeElement;
    const container = containerRef.current;
    if (!container) return;
    const focusables = getFocusables(container);
    if (focusables.length) focusables[0].focus();

    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
        return;
      }
      if (e.key !== 'Tab') return;
      const list = getFocusables(container);
      if (list.length === 0) return;
      const first = list[0];
      const last = list[list.length - 1];
      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    container.addEventListener('keydown', handleKeyDown);
    return () => {
      container.removeEventListener('keydown', handleKeyDown);
      if (previousActiveRef.current && typeof previousActiveRef.current.focus === 'function') {
        previousActiveRef.current.focus();
      }
    };
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      role="dialog"
      aria-modal="true"
      aria-labelledby="api-key-modal-title"
      aria-describedby="api-key-modal-desc"
    >
      <div
        ref={containerRef}
        className="p-8 bg-white rounded-xl shadow-2xl w-full max-w-lg border border-gray-200"
        onClick={(e) => e.stopPropagation()}
      >
        <h1 id="api-key-modal-title" className="text-2xl font-bold mb-4 text-gray-900">
          Sua Chave de API foi Gerada
        </h1>
        <p id="api-key-modal-desc" className="text-gray-600 mb-4">
          Esta é a sua chave de API. Ela é usada para autenticar suas requisições aos nossos serviços.
          <span className="inline-flex ml-1 relative group">
            <InfoIcon />
            <span className="absolute bottom-full mb-2 w-64 p-2 bg-gray-800 text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none">
              A chave de API permite que seus sistemas se comuniquem de forma segura com a plataforma NXD para enviar e receber dados.
            </span>
          </span>
        </p>

        <div className="bg-amber-50 border-l-4 border-amber-500 text-amber-800 p-4 mb-4 rounded-r-lg">
          <p className="font-bold">Atenção: Anote sua chave em um lugar seguro.</p>
          <p>
            Por motivos de segurança, esta é a <strong>única vez</strong> que a chave será exibida. Ao fechar esta janela, não será possível recuperá-la.
          </p>
        </div>

        <div className="relative mb-4">
          <input
            type="text"
            readOnly
            value={apiKey}
            className="w-full px-3 py-2 bg-gray-100 border border-gray-300 rounded-lg pr-12"
            aria-label="Chave de API gerada"
          />
          {apiKey && (
            <button
              type="button"
              onClick={handleCopy}
              className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500 hover:text-gray-800 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-inset rounded-r-lg"
              aria-label="Copiar chave"
            >
              <CopyIcon />
            </button>
          )}
        </div>
        {copySuccess && (
          <p className="text-green-600 text-sm mb-4" role="status" aria-live="polite">
            {copySuccess}
          </p>
        )}

        <div className="flex items-start gap-3 mb-6">
          <input
            type="checkbox"
            id="ack-check"
            className="mt-1 h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            checked={hasAcknowledged}
            onChange={(e) => setHasAcknowledged(e.target.checked)}
            aria-describedby="ack-label"
          />
          <label id="ack-label" htmlFor="ack-check" className="text-sm text-gray-700 flex-1">
            Entendo que preciso salvar esta chave agora e que não poderei vê-la novamente.
          </label>
        </div>

        <div className="flex justify-end">
          <button
            type="button"
            onClick={onClose}
            disabled={!hasAcknowledged}
            className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-5 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
          >
            Concluir e Ir para o Sistema
          </button>
        </div>
      </div>
    </div>
  );
}
