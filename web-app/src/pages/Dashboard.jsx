import React, { useState } from 'react';

export default function Dashboard() {
  const [apiKey, setApiKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleGenerateKey = async () => {
    setLoading(true);
    setError('');
    try {
      const token = localStorage.getItem('nxd-token');
      if (!token) {
        throw new Error('Você não está autenticado.');
      }

      const response = await fetch('/api/factory/generate-key', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Falha ao gerar a chave.');
      }

      const { apiKey } = await response.json();
      setApiKey(apiKey);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
      <div className="p-6 bg-white rounded-lg shadow-md">
        <h2 className="text-xl font-semibold mb-2">Sua Chave de API</h2>
        <p className="text-gray-600 mb-4">
          Use esta chave para se conectar aos serviços do NXD.
        </p>
        <div className="flex items-center space-x-4">
          <input
            type="text"
            readOnly
            value={apiKey || 'Nenhuma chave gerada ainda.'}
            className="w-full px-3 py-2 bg-gray-100 border rounded-lg focus:outline-none"
          />
          <button
            onClick={handleGenerateKey}
            disabled={loading}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg disabled:bg-gray-400"
          >
            {loading ? 'Gerando...' : 'Gerar Nova Chave'}
          </button>
        </div>
        {apiKey && (
          <p className="text-sm text-green-600 mt-2">Chave gerada com sucesso!</p>
        )}
        {error && <p className="text-red-500 text-xs italic mt-2">{error}</p>}
      </div>
    </div>
  );
}
