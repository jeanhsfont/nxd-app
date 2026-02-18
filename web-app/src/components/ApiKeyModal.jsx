import React, { useState } from 'react';

// Ícone de Copiar (Simples, em SVG)
const CopyIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
  </svg>
);

// Ícone de Informação (Simples, em SVG)
const InfoIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
);


export default function ApiKeyModal({ apiKey, onClose }) {
  const [hasAcknowledged, setHasAcknowledged] = useState(false);
  const [copySuccess, setCopySuccess] = useState('');

  const handleCopy = () => {
    navigator.clipboard.writeText(apiKey);
    setCopySuccess('Chave copiada para a área de transferência!');
    setTimeout(() => setCopySuccess(''), 2000);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="p-8 bg-white rounded-lg shadow-2xl w-full max-w-lg">
        <h1 className="text-2xl font-bold mb-4">Sua Chave de API foi Gerada</h1>
        <p className="text-gray-600 mb-4">
          Esta é a sua chave de API. Ela é usada para autenticar suas requisições aos nossos serviços.
          <span className="inline-flex ml-1 relative group">
            <InfoIcon />
            <span className="absolute bottom-full mb-2 w-64 p-2 bg-gray-800 text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity duration-300">
              A chave de API permite que seus sistemas se comuniquem de forma segura com a plataforma NXD para enviar e receber dados.
            </span>
          </span>
        </p>
        
        <div className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 mb-4 rounded-md">
          <p className="font-bold">Atenção: Anote sua chave em um lugar seguro.</p>
          <p>Por motivos de segurança, esta é a **única vez** que a chave será exibida. Ao fechar esta janela, não será possível recuperá-la.</p>
        </div>

        <div className="relative mb-4">
          <input
            type="text"
            readOnly
            value={apiKey}
            className="w-full px-3 py-2 bg-gray-100 border rounded-lg pr-10"
          />
          {apiKey && (
            <button onClick={handleCopy} className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500 hover:text-gray-800">
              <CopyIcon />
            </button>
          )}
        </div>
        {copySuccess && <p className="text-green-600 text-sm mb-4">{copySuccess}</p>}

        <div className="flex items-center mb-6">
          <input type="checkbox" id="ack-check" className="mr-2" checked={hasAcknowledged} onChange={(e) => setHasAcknowledged(e.target.checked)} />
          <label htmlFor="ack-check" className="text-sm text-gray-700">
            Entendo que preciso salvar esta chave agora e que não poderei vê-la novamente.
          </label>
        </div>

        <div className="flex justify-end items-center">
          <button
            onClick={onClose}
            disabled={!hasAcknowledged}
            className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded-lg disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            Concluir e Ir para o Sistema
          </button>
        </div>
      </div>
    </div>
  );
}
