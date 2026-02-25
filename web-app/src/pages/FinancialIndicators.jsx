import React, { useState, useEffect } from 'react';
import {
  DollarSign,
  TrendingUp,
  Clock,
  RefreshCw,
  BarChart2,
  Plus,
  Save,
  FileDown,
  FileSpreadsheet,
} from 'lucide-react';
import { jsPDF } from 'jspdf';
import api from '../utils/api';

function formatCurrency(v) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v || 0);
}

function SummaryCards({ ranges, loading }) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="nxd-card h-32 animate-pulse bg-gray-100" />
        ))}
      </div>
    );
  }
  const today = ranges?.today;
  const last24h = ranges?.['24h'];
  const last7d = ranges?.['7d'];
  const last30d = ranges?.['30d'];
  const cards = [
    { key: 'today', label: 'Hoje', icon: Clock, data: today },
    { key: '24h', label: 'Últimas 24h', icon: BarChart2, data: last24h },
    { key: '7d', label: 'Últimos 7 dias', icon: TrendingUp, data: last7d },
    { key: '30d', label: 'Últimos 30 dias', icon: TrendingUp, data: last30d },
  ];
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {cards.map(({ key, label, icon: Icon, data }) => (
        <div key={key} className="nxd-card">
          <div className="flex items-center gap-2 text-gray-500 text-sm mb-3">
            <Icon className="w-4 h-4" />
            {label}
          </div>
          <div className="space-y-2">
            <p className="text-xs text-gray-500">Faturamento bruto</p>
            <p className="text-xl font-bold text-green">{formatCurrency(data?.faturamento_bruto)}</p>
            <p className="text-xs text-gray-500">Perda refugo / Custo parada</p>
            <p className="text-sm text-red">{formatCurrency(data?.perda_refugo)} / {formatCurrency(data?.custo_parada)}</p>
          </div>
        </div>
      ))}
    </div>
  );
}

export default function FinancialIndicators() {
  const [ranges, setRanges] = useState(null);
  const [exportComparativo, setExportComparativo] = useState({ '7d': null, '30d': null });
  const [configs, setConfigs] = useState([]);
  const [mappings, setMappings] = useState([]);
  const [sectors, setSectors] = useState([]);
  const [assets, setAssets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [configOpen, setConfigOpen] = useState(false);
  const [mappingOpen, setMappingOpen] = useState(false);
  const [selectedSectorId, setSelectedSectorId] = useState('');
  const [configForm, setConfigForm] = useState({
    sector_id: '',
    valor_venda_ok: '',
    custo_refugo_un: '',
    custo_parada_h: '',
  });
  const [mappingForm, setMappingForm] = useState({
    asset_id: '',
    tag_ok: '',
    tag_nok: '',
    tag_status: '',
    reading_rule: 'delta',
  });

  const fetchData = async () => {
    setLoading(true);
    const params = selectedSectorId ? { sector_id: selectedSectorId } : {};
    try {
      const [rangesRes, configsRes, mappingsRes, sectorsRes, assetsRes, export7Res, export30Res] = await Promise.all([
        api.get('/api/financial-summary/ranges', { params }),
        api.get('/api/business-config'),
        api.get('/api/tag-mappings'),
        api.get('/api/sectors'),
        api.get('/api/dashboard/data'),
        api.get('/api/financial-summary/export', { params: { ...params, period: '7d' } }).catch(() => ({ data: null })),
        api.get('/api/financial-summary/export', { params: { ...params, period: '30d' } }).catch(() => ({ data: null })),
      ]);
      setRanges(rangesRes.data);
      setExportComparativo({ '7d': export7Res?.data || null, '30d': export30Res?.data || null });
      setConfigs(configsRes.data.configs || []);
      setMappings(mappingsRes.data.mappings || []);
      setSectors(Array.isArray(sectorsRes.data) ? sectorsRes.data : (sectorsRes.data?.sectors || []));
      setAssets(assetsRes.data.assets || []);
    } catch (err) {
      console.error('Erro ao carregar dados financeiros:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedSectorId]);

  const handleSaveConfig = async () => {
    try {
      await api.post('/api/business-config', configForm);
      await fetchData();
      setConfigOpen(false);
      setConfigForm({ sector_id: '', valor_venda_ok: '', custo_refugo_un: '', custo_parada_h: '' });
    } catch (err) {
      console.error('Erro ao salvar configuração:', err);
    }
  };

  const handleSaveMapping = async () => {
    try {
      await api.post('/api/tag-mappings', mappingForm);
      await fetchData();
      setMappingOpen(false);
      setMappingForm({ asset_id: '', tag_ok: '', tag_nok: '', tag_status: '', reading_rule: 'delta' });
    } catch (err) {
      console.error('Erro ao salvar mapeamento:', err);
    }
  };

  const exportCSV = () => {
    if (!ranges) return;
    const sep = ';';
    const header = 'Período;Faturamento bruto;Perda refugo;Custo parada;Perdas evitadas;Custo parada evitado';
    const row = (label, d, pe, cpe) =>
      [label, (d?.faturamento_bruto ?? ''), (d?.perda_refugo ?? ''), (d?.custo_parada ?? ''), pe ?? '', cpe ?? ''].join(sep);
    const e7 = exportComparativo['7d'];
    const e30 = exportComparativo['30d'];
    const rows = [
      row('Hoje', ranges.today, '', ''),
      row('Últimas 24h', ranges['24h'], '', ''),
      row('Últimos 7 dias', ranges['7d'], e7?.perdas_evitadas ?? '', e7?.custo_parada_evitado ?? ''),
      row('Últimos 30 dias', ranges['30d'], e30?.perdas_evitadas ?? '', e30?.custo_parada_evitado ?? ''),
    ];
    const csv = [header, ...rows].map((r) => r.replace(/\./g, ',')).join('\r\n');
    const blob = new Blob(['\ufeff' + csv], { type: 'text/csv;charset=utf-8' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `resumo-financeiro-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(a.href);
  };

  const exportPDF = () => {
    if (!ranges) return;
    const doc = new jsPDF();
    doc.setFontSize(16);
    doc.text('Resumo executivo – Indicadores financeiros', 14, 20);
    doc.setFontSize(10);
    doc.text(new Date().toLocaleDateString('pt-BR'), 14, 28);
    const e7 = exportComparativo['7d'];
    const e30 = exportComparativo['30d'];
    const lineH = 7;
    let y = 38;
    doc.setFont(undefined, 'bold');
    doc.text('Período', 14, y);
    doc.text('Faturamento', 50, y);
    doc.text('Perda refugo', 85, y);
    doc.text('Custo parada', 120, y);
    doc.text('Perdas evitadas', 155, y);
    doc.text('Custo parada evit.', 190, y);
    y += lineH;
    doc.setFont(undefined, 'normal');
    const addRow = (label, d, pe, cpe) => {
      doc.text(label, 14, y);
      doc.text(formatCurrency(d?.faturamento_bruto), 50, y);
      doc.text(formatCurrency(d?.perda_refugo), 85, y);
      doc.text(formatCurrency(d?.custo_parada), 120, y);
      doc.text(pe != null ? formatCurrency(pe) : '-', 155, y);
      doc.text(cpe != null ? formatCurrency(cpe) : '-', 190, y);
      y += lineH;
    };
    addRow('Hoje', ranges.today, null, null);
    addRow('Últimas 24h', ranges['24h'], null, null);
    addRow('Últimos 7 dias', ranges['7d'], e7?.perdas_evitadas, e7?.custo_parada_evitado);
    addRow('Últimos 30 dias', ranges['30d'], e30?.perdas_evitadas, e30?.custo_parada_evitado);
    doc.save(`resumo-executivo-${new Date().toISOString().slice(0, 10)}.pdf`);
  };

  const filteredAssets = selectedSectorId
    ? assets.filter((a) => a.group_id === selectedSectorId)
    : assets;

  const filteredConfigs = selectedSectorId
    ? configs.filter((c) => c.sector_id === selectedSectorId)
    : configs;

  const filteredMappings = selectedSectorId
    ? mappings.filter((m) => {
        const asset = assets.find((a) => a.id === m.asset_id);
        return asset && asset.group_id === selectedSectorId;
      })
    : mappings;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto p-6">
        {/* Header */}
        <div className="page-header">
          <div className="page-header-icon">
            <DollarSign className="w-6 h-6" />
          </div>
          <div className="flex-1">
            <h1 className="page-title">Indicadores Financeiros</h1>
            <p className="page-subtitle">Análise de faturamento, perdas e custos</p>
          </div>
          <div className="flex items-center gap-2">
            <button onClick={exportCSV} disabled={!ranges} className="nxd-btn nxd-btn-secondary" title="Exportar CSV">
              <FileSpreadsheet className="w-5 h-5" />
              CSV
            </button>
            <button onClick={exportPDF} disabled={!ranges} className="nxd-btn nxd-btn-secondary" title="Exportar PDF">
              <FileDown className="w-5 h-5" />
              PDF
            </button>
            <button onClick={fetchData} className="nxd-btn nxd-btn-primary">
              <RefreshCw className="w-5 h-5" />
              Atualizar
            </button>
          </div>
        </div>

        {/* Summary Cards */}
        <SummaryCards ranges={ranges} loading={loading} />

        {/* Comparativo: perdas evitadas e custo parada evitado */}
        {(exportComparativo['7d'] || exportComparativo['30d']) && (
          <div className="nxd-card mt-6">
            <h2 className="text-lg font-bold text-gray-900 mb-4">Comparativo com período anterior</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {exportComparativo['7d'] && (
                <div>
                  <p className="text-sm text-gray-500 mb-2">Últimos 7 dias</p>
                  <p className="text-sm"><span className="text-gray-600">Perdas evitadas:</span> <span className="font-semibold text-green">{formatCurrency(exportComparativo['7d'].perdas_evitadas)}</span></p>
                  <p className="text-sm"><span className="text-gray-600">Custo de parada evitado:</span> <span className="font-semibold text-green">{formatCurrency(exportComparativo['7d'].custo_parada_evitado)}</span></p>
                </div>
              )}
              {exportComparativo['30d'] && (
                <div>
                  <p className="text-sm text-gray-500 mb-2">Últimos 30 dias</p>
                  <p className="text-sm"><span className="text-gray-600">Perdas evitadas:</span> <span className="font-semibold text-green">{formatCurrency(exportComparativo['30d'].perdas_evitadas)}</span></p>
                  <p className="text-sm"><span className="text-gray-600">Custo de parada evitado:</span> <span className="font-semibold text-green">{formatCurrency(exportComparativo['30d'].custo_parada_evitado)}</span></p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Filter */}
        <div className="nxd-card mt-6">
          <label className="block text-sm font-medium text-gray-700 mb-2">Filtrar por Setor</label>
          <select
            value={selectedSectorId}
            onChange={(e) => setSelectedSectorId(e.target.value)}
            className="nxd-input"
          >
            <option value="">Todos os Setores</option>
            {sectors.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </div>

        {/* Configurations */}
        <div className="nxd-card mt-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-gray-900">Configurações de Setor</h2>
            <button onClick={() => setConfigOpen(!configOpen)} className="nxd-btn nxd-btn-primary">
              <Plus className="w-5 h-5" />
              {configOpen ? 'Cancelar' : 'Adicionar'}
            </button>
          </div>

          {configOpen && (
            <div className="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200 space-y-3">
              <select
                value={configForm.sector_id}
                onChange={(e) => setConfigForm({ ...configForm, sector_id: e.target.value })}
                className="nxd-input"
              >
                <option value="">Selecione um setor</option>
                {sectors.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
              <input
                type="number"
                placeholder="Valor venda OK (R$)"
                value={configForm.valor_venda_ok}
                onChange={(e) => setConfigForm({ ...configForm, valor_venda_ok: e.target.value })}
                className="nxd-input"
              />
              <input
                type="number"
                placeholder="Custo refugo/un (R$)"
                value={configForm.custo_refugo_un}
                onChange={(e) => setConfigForm({ ...configForm, custo_refugo_un: e.target.value })}
                className="nxd-input"
              />
              <input
                type="number"
                placeholder="Custo parada/h (R$)"
                value={configForm.custo_parada_h}
                onChange={(e) => setConfigForm({ ...configForm, custo_parada_h: e.target.value })}
                className="nxd-input"
              />
              <button onClick={handleSaveConfig} className="nxd-btn nxd-btn-primary w-full justify-center">
                <Save className="w-5 h-5" />
                Salvar Configuração
              </button>
            </div>
          )}

          {filteredConfigs.length === 0 ? (
            <p className="text-gray-500 text-sm text-center py-8">Nenhuma configuração cadastrada</p>
          ) : (
            <div className="space-y-2">
              {filteredConfigs.map((cfg) => {
                const sector = sectors.find((s) => s.id === cfg.sector_id);
                return (
                  <div key={cfg.id} className="p-3 bg-gray-50 rounded-lg border border-gray-200">
                    <p className="font-semibold text-gray-900">{sector?.name || 'Setor desconhecido'}</p>
                    <div className="grid grid-cols-3 gap-2 mt-2 text-sm text-gray-600">
                      <div>
                        <span className="text-xs text-gray-500">Valor OK:</span>
                        <p className="font-medium">{formatCurrency(cfg.valor_venda_ok)}</p>
                      </div>
                      <div>
                        <span className="text-xs text-gray-500">Custo Refugo:</span>
                        <p className="font-medium">{formatCurrency(cfg.custo_refugo_un)}</p>
                      </div>
                      <div>
                        <span className="text-xs text-gray-500">Custo Parada/h:</span>
                        <p className="font-medium">{formatCurrency(cfg.custo_parada_h)}</p>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Mappings */}
        <div className="nxd-card mt-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-gray-900">Mapeamento de Tags</h2>
            <button onClick={() => setMappingOpen(!mappingOpen)} className="nxd-btn nxd-btn-primary">
              <Plus className="w-5 h-5" />
              {mappingOpen ? 'Cancelar' : 'Adicionar'}
            </button>
          </div>

          {mappingOpen && (
            <div className="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200 space-y-3">
              <select
                value={mappingForm.asset_id}
                onChange={(e) => setMappingForm({ ...mappingForm, asset_id: e.target.value })}
                className="nxd-input"
              >
                <option value="">Selecione um ativo</option>
                {filteredAssets.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.display_name}
                  </option>
                ))}
              </select>
              <input
                type="text"
                placeholder="Tag OK (ex: Total_Pecas)"
                value={mappingForm.tag_ok}
                onChange={(e) => setMappingForm({ ...mappingForm, tag_ok: e.target.value })}
                className="nxd-input"
              />
              <input
                type="text"
                placeholder="Tag NOK (ex: Fault_Code)"
                value={mappingForm.tag_nok}
                onChange={(e) => setMappingForm({ ...mappingForm, tag_nok: e.target.value })}
                className="nxd-input"
              />
              <input
                type="text"
                placeholder="Tag Status (ex: Status_Producao)"
                value={mappingForm.tag_status}
                onChange={(e) => setMappingForm({ ...mappingForm, tag_status: e.target.value })}
                className="nxd-input"
              />
              <button onClick={handleSaveMapping} className="nxd-btn nxd-btn-primary w-full justify-center">
                <Save className="w-5 h-5" />
                Salvar Mapeamento
              </button>
            </div>
          )}

          {filteredMappings.length === 0 ? (
            <p className="text-gray-500 text-sm text-center py-8">Nenhum mapeamento cadastrado</p>
          ) : (
            <div className="space-y-2">
              {filteredMappings.map((map) => {
                const asset = assets.find((a) => a.id === map.asset_id);
                return (
                  <div key={map.id} className="p-3 bg-gray-50 rounded-lg border border-gray-200">
                    <p className="font-semibold text-gray-900">{asset?.display_name || 'Ativo desconhecido'}</p>
                    <div className="grid grid-cols-3 gap-2 mt-2 text-xs text-gray-600">
                      <div>
                        <span className="text-gray-500">Tag OK:</span>
                        <p className="font-mono">{map.tag_ok}</p>
                      </div>
                      <div>
                        <span className="text-gray-500">Tag NOK:</span>
                        <p className="font-mono">{map.tag_nok}</p>
                      </div>
                      <div>
                        <span className="text-gray-500">Tag Status:</span>
                        <p className="font-mono">{map.tag_status}</p>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
