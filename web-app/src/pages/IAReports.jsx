import React, { useState, useEffect } from 'react';
import { FileText, TrendingUp, Loader2 } from 'lucide-react';
import api from '../utils/api';

const IAReports = () => {
  const [sectors, setSectors] = useState([]);
  const [selectedSector, setSelectedSector] = useState('');
  const [reportData, setReportData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const [reports, setReports] = useState([]);
  const [selectedReportId, setSelectedReportId] = useState(null);
  const [reportDetail, setReportDetail] = useState(null);

  useEffect(() => {
    const fetchSectors = async () => {
      try {
        const response = await api.get('/api/sectors');
        setSectors(Array.isArray(response.data) ? response.data : (response.data?.sectors || []));
      } catch (err) {
        setError('Não foi possível carregar os setores.');
        console.error(err);
      }
    };
    fetchSectors();
  }, []);

  useEffect(() => {
    api.get('/api/ia/reports').then((res) => setReports(res.data.reports || [])).catch(() => {});
  }, [reportData]);

  const handleGenerateReport = async (executivo = false) => {
    setLoading(true);
    setError('');
    setReportData(null);
    try {
      const url = executivo ? '/api/ia/analysis' : `/api/ia/analysis?sector_id=${selectedSector}`;
      const response = await api.get(url);
      setReportData(response.data);
    } catch (err) {
      setError('Não foi possível gerar o relatório.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenReport = (id) => {
    setSelectedReportId(id);
    api.get(`/api/ia/reports/${id}`).then((res) => setReportDetail(res.data)).catch(() => setReportDetail(null));
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6">
        <div className="page-header">
          <div className="page-header-icon">
            <TrendingUp className="w-6 h-6" />
          </div>
          <div>
            <h1 className="page-title">Relatórios com IA</h1>
            <p className="page-subtitle">Análises inteligentes dos seus dados industriais</p>
          </div>
        </div>

        <div className="nxd-card mb-6">
          <div className="mb-5">
            <label htmlFor="sector-select" className="block text-sm font-medium text-gray-700 mb-2">
              Selecione um Setor para Análise
            </label>
            <select
              id="sector-select"
              value={selectedSector}
              onChange={(e) => setSelectedSector(e.target.value)}
              className="nxd-input"
            >
              <option value="">-- Escolha um setor --</option>
              {sectors.map((sector) => (
                <option key={sector.id} value={sector.id}>
                  {sector.name}
                </option>
              ))}
            </select>
          </div>

          <div className="flex gap-2">
            <button
              onClick={() => handleGenerateReport(false)}
              disabled={loading || !selectedSector}
              className="nxd-btn nxd-btn-primary"
            >
              {loading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  Gerando...
                </>
              ) : (
                <>
                  <FileText className="w-5 h-5" />
                  Gerar por setor
                </>
              )}
            </button>
            <button
              onClick={() => handleGenerateReport(true)}
              disabled={loading}
              className="nxd-btn nxd-btn-secondary"
            >
              <FileText className="w-5 h-5" />
              Relatório executivo
            </button>
          </div>

          {error && <p className="text-red text-sm mt-4">{error}</p>}
        </div>

        {reportData && (
          <div className="nxd-card fade-in">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Resultados da Análise</h2>
            {reportData.sources && (
              <p className="text-xs text-gray-500 mb-2">Baseado em: {reportData.sources}</p>
            )}
            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
              <p className="whitespace-pre-wrap text-gray-700">{reportData.analysis}</p>
            </div>
          </div>
        )}

        {reports.length > 0 && (
          <div className="nxd-card mt-6">
            <h2 className="text-lg font-bold text-gray-900 mb-4">Histórico de análises</h2>
            <ul className="space-y-2">
              {reports.map((r) => (
                <li key={r.id}>
                  <button
                    type="button"
                    onClick={() => handleOpenReport(r.id)}
                    className="text-left w-full px-3 py-2 rounded-lg hover:bg-gray-100 text-gray-800"
                  >
                    <span className="font-medium">{r.title}</span>
                    <span className="text-xs text-gray-500 ml-2">{new Date(r.created_at).toLocaleString('pt-BR')}</span>
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}

        {reportDetail && (
          <div className="nxd-card mt-6 fade-in">
            <h2 className="text-xl font-bold text-gray-900 mb-4">{reportDetail.title}</h2>
            {reportDetail.sources && (
              <p className="text-xs text-gray-500 mb-2">Baseado em: {reportDetail.sources}</p>
            )}
            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
              <p className="whitespace-pre-wrap text-gray-700">{reportDetail.analysis}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default IAReports;
