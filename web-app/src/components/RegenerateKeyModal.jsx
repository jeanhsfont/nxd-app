import React, { useState } from 'react';

const CopyIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
  </svg>
);

export default function RegenerateKeyModal({ onClose }) {
  const [newApiKey, setNewApiKey] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [hasAcknowledged, setHasAcknowledged] = useState(false);
  const [copySuccess, setCopySuccess] = useState('');
  const [error, setError] = useState('');

  const handleRegenerateKey = async () => {
    setIsLoading(true);
    setError('');
    setCopySuccess('');
    try {
      const token = localStorage.getItem('nxd-token');
      const response = await fetch('/api/factory/regenerate-api-key', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
      });
      if (!response.ok) {
        throw new Error('Falha ao gerar a nova chave. Tente novamente.');
      }
      const data = await response.json();
      setNewApiKey(data.api_key);
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(newApiKey);
    setCopySuccess('Chave copiada para a área de transferência!');
    setTimeout(() => setCopySuccess(''), 2000);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" role="dialog" aria-modal="true" aria-labelledby="regen-modal-title">
      <div className="p-8 bg-white rounded-xl shadow-2xl w-full max-w-lg border border-gray-200">
        <h1 id="regen-modal-title" className="text-2xl font-bold mb-4 text-red-600">Atenção: Regenerar Chave de API</h1>
        <p className="text-gray-600 mb-4">
          Você está prestes a invalidar sua chave de API atual. Todos os sistemas que a utilizam deixarão de funcionar até que sejam atualizados com a nova chave.
        </p>
        
        {newApiKey ? (
          <>
            <div className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 mb-4 rounded-md">
              <p className="font-bold">Esta é a sua nova chave. Anote-a em um lugar seguro.</p>
              <p>Por motivos de segurança, esta é a única vez que a chave será exibida.</p>
            </div>
            <div className="relative mb-4">
              <input type="text" readOnly value={newApiKey} className="w-full px-3 py-2 bg-gray-100 border rounded-lg pr-10" />
              <button type="button" onClick={handleCopy} className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500 hover:text-gray-800" aria-label="Copiar chave">
                <CopyIcon />
              </button>
            </div>
            {copySuccess && <p className="text-green-600 text-sm mb-4" role="status" aria-live="polite">{copySuccess}</p>}
            <div className="flex justify-end">
              <button type="button" onClick={onClose} className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-4 rounded-lg">
                Fechar
              </button>
            </div>
          </>
        ) : (
          <>
            <div className="flex items-center mb-6">
              <input type="checkbox" id="ack-regen-check" className="mr-2" checked={hasAcknowledged} onChange={(e) => setHasAcknowledged(e.target.checked)} />
              <label htmlFor="ack-regen-check" className="text-sm text-gray-700">
                Eu entendo que minha chave de API atual será permanentemente invalidada e que esta nova chave não poderá ser vista novamente após fechar esta janela.
              </label>
            </div>
            {error && <p className="text-red-600 text-sm mb-4 p-2 bg-red-50 border border-red-200 rounded-lg" role="alert">{error}</p>}
            <div className="flex justify-between items-center gap-3">
              <button type="button" onClick={onClose} className="bg-gray-500 hover:bg-gray-700 text-white font-semibold py-2 px-4 rounded-lg">
                Cancelar
              </button>
              <button
                type="button"
                onClick={handleRegenerateKey}
                disabled={!hasAcknowledged || isLoading}
                className="bg-red-600 hover:bg-red-700 text-white font-semibold py-2 px-4 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Gerando...' : 'Invalidar Chave Antiga e Gerar Nova'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
