import React, { useState, useEffect } from 'react';
import api from '../services/api';

const IAReports = () => {
  const [sectors, setSectors] = useState([]);
  const [selectedSector, setSelectedSector] = useState('');
  const [reportData, setReportData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchSectors = async () => {
      try {
        const response = await api.get('/api/sectors');
        setSectors(response.data);
      } catch (err) {
        setError('Não foi possível carregar os setores.');
        console.error(err);
      }
    };
    fetchSectors();
  }, []);

  const handleGenerateReport = async () => {
    if (!selectedSector) {
      setError('Por favor, selecione um setor.');
      return;
    }
    setLoading(true);
    setError('');
    setReportData(null);
    try {
      const response = await api.get(`/api/ia/analysis?sector_id=${selectedSector}`);
      setReportData(response.data);
    } catch (err) {
      setError('Não foi possível gerar o relatório.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Relatórios com Inteligência Artificial</h1>

      <div className="bg-white p-6 rounded-lg shadow-md">
        <div className="mb-4">
          <label htmlFor="sector-select" className="block text-sm font-medium text-gray-700 mb-2">
            Selecione um Setor para Análise
          </label>
          <select
            id="sector-select"
            value={selectedSector}
            onChange={(e) => setSelectedSector(e.target.value)}
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          >
            <option value="">-- Escolha um setor --</option>
            {sectors.map((sector) => (
              <option key={sector.id} value={sector.id}>
                {sector.name}
              </option>
            ))}
          </select>
        </div>

        <button
          onClick={handleGenerateReport}
          disabled={loading || !selectedSector}
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded disabled:bg-gray-400"
        >
          {loading ? 'Gerando...' : 'Gerar Relatório'}
        </button>

        {error && <p className="text-red-500 mt-4">{error}</p>}
      </div>

      {reportData && (
        <div className="mt-6 bg-white p-6 rounded-lg shadow-md">
          <h2 className="text-xl font-bold mb-4">Resultados da Análise</h2>
          <pre className="bg-gray-100 p-4 rounded-md overflow-x-auto">
            {JSON.stringify(reportData, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
};

export default IAReports;
