import React, { useState, useEffect } from 'react';
import {
  DollarSign,
  TrendingUp,
  AlertTriangle,
  Clock,
  Settings,
  RefreshCw,
  BarChart2,
  Tag,
  Loader2,
} from 'lucide-react';
import api from '../utils/api';

function formatCurrency(v) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v || 0);
}

function SummaryCards({ ranges, loading }) {
  if (loading) {
    return (
      <div className="flex gap-4 flex-wrap">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex-1 min-w-[200px] h-28 bg-gray-100 rounded-xl animate-pulse" />
        ))}
      </div>
    );
  }
  const today = ranges?.today;
  const last24h = ranges?.['24h'];
  const last7d = ranges?.['7d'];
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
        <div className="flex items-center gap-2 text-gray-500 text-sm mb-2">
          <Clock className="w-4 h-4" />
          Hoje
        </div>
        <div className="space-y-1">
          <p className="text-xs text-gray-500">Faturamento bruto</p>
          <p className="text-lg font-bold text-emerald-600">{formatCurrency(today?.faturamento_bruto)}</p>
          <p className="text-xs text-gray-500">Perda refugo / Custo parada</p>
          <p className="text-sm text-red-600">{formatCurrency(today?.perda_refugo)} / {formatCurrency(today?.custo_parada)}</p>
        </div>
      </div>
      <div className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
        <div className="flex items-center gap-2 text-gray-500 text-sm mb-2">
          <BarChart2 className="w-4 h-4" />
          Últimas 24h
        </div>
        <div className="space-y-1">
          <p className="text-xs text-gray-500">Faturamento bruto</p>
          <p className="text-lg font-bold text-emerald-600">{formatCurrency(last24h?.faturamento_bruto)}</p>
          <p className="text-xs text-gray-500">Perda refugo / Custo parada</p>
          <p className="text-sm text-red-600">{formatCurrency(last24h?.perda_refugo)} / {formatCurrency(last24h?.custo_parada)}</p>
        </div>
      </div>
      <div className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
        <div className="flex items-center gap-2 text-gray-500 text-sm mb-2">
          <TrendingUp className="w-4 h-4" />
          Últimos 7 dias
        </div>
        <div className="space-y-1">
          <p className="text-xs text-gray-500">Faturamento bruto</p>
          <p className="text-lg font-bold text-emerald-600">{formatCurrency(last7d?.faturamento_bruto)}</p>
          <p className="text-xs text-gray-500">Perda refugo / Custo parada</p>
          <p className="text-sm text-red-600">{formatCurrency(last7d?.perda_refugo)} / {formatCurrency(last7d?.custo_parada)}</p>
        </div>
      </div>
    </div>
  );
}

export default function FinancialIndicators() {
  const [ranges, setRanges] = useState(null);
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
    try {
      const sectorQ = selectedSectorId ? `?sector_id=${selectedSectorId}` : '';
      const [rangesRes, configsRes, mappingsRes, sectorsRes, dashRes] = await Promise.all([
        api.get(`/api/financial-summary/ranges${sectorQ}`),
        api.get('/api/business-config'),
        api.get('/api/tag-mappings'),
        api.get('/api/sectors'),
        api.get('/api/dashboard/data').catch(() => ({ data: {} })),
      ]);
      setRanges(rangesRes.data);
      setConfigs(configsRes.data?.configs || []);
      setMappings(mappingsRes.data?.mappings || []);
      setSectors(Array.isArray(sectorsRes.data) ? sectorsRes.data : (sectorsRes.data?.sectors || []));
      setAssets(dashRes.data?.assets || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [selectedSectorId]);

  const saveConfig = async () => {
    try {
      await api.post('/api/business-config', {
        sector_id: configForm.sector_id || null,
        valor_venda_ok: parseFloat(configForm.valor_venda_ok) || 0,
        custo_refugo_un: parseFloat(configForm.custo_refugo_un) || 0,
        custo_parada_h: parseFloat(configForm.custo_parada_h) || 0,
      });
      setConfigOpen(false);
      fetchData();
    } catch (e) {
      console.error(e);
    }
  };

  const saveMapping = async () => {
    if (!mappingForm.asset_id) return;
    try {
      await api.post('/api/tag-mappings', {
        asset_id: mappingForm.asset_id,
        tag_ok: mappingForm.tag_ok,
        tag_nok: mappingForm.tag_nok,
        tag_status: mappingForm.tag_status,
        reading_rule: mappingForm.reading_rule,
      });
      setMappingOpen(false);
      setMappingForm({ asset_id: '', tag_ok: '', tag_nok: '', tag_status: '', reading_rule: 'delta' });
      fetchData();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="p-8 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <DollarSign className="w-8 h-8 text-emerald-600" />
            Indicadores financeiros
          </h1>
          <p className="text-gray-500 mt-1">
            Faturamento bruto, perda por refugo e custo de parada por setor/linha. Configure os parâmetros e o mapeamento de tags do CLP.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={selectedSectorId}
            onChange={(e) => setSelectedSectorId(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-2 text-sm"
          >
            <option value="">Toda a fábrica</option>
            {sectors.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
          <button type="button" onClick={fetchData} className="p-2 text-gray-500 hover:bg-gray-100 rounded-lg">
            <RefreshCw className="w-5 h-5" />
          </button>
          <button
            type="button"
            onClick={() => setConfigOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-gray-800 text-white rounded-lg hover:bg-gray-700"
          >
            <Settings className="w-5 h-5" />
            Config. negócio
          </button>
          <button
            type="button"
            onClick={() => setMappingOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
          >
            <Tag className="w-5 h-5" />
            Mapear tags
          </button>
        </div>
      </div>

      <section className="mb-8">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Resumo por período</h2>
        <SummaryCards ranges={ranges} loading={loading} />
      </section>

      <section>
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Configurações de negócio</h2>
        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
          {configs.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              Nenhuma configuração. Clique em &quot;Config. negócio&quot; e defina valor de venda (R$/un), custo refugo (R$/un) e custo hora parada (R$/h).
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="text-left p-3">Setor</th>
                  <th className="text-right p-3">Valor venda OK (R$/un)</th>
                  <th className="text-right p-3">Custo refugo (R$/un)</th>
                  <th className="text-right p-3">Custo parada (R$/h)</th>
                </tr>
              </thead>
              <tbody>
                {configs.map((c) => (
                  <tr key={c.id} className="border-b border-gray-100">
                    <td className="p-3">{c.sector_id ? sectors.find(s => s.id === c.sector_id)?.name || c.sector_id : 'Padrão (fábrica)'}</td>
                    <td className="p-3 text-right">{formatCurrency(c.valor_venda_ok)}</td>
                    <td className="p-3 text-right">{formatCurrency(c.custo_refugo_un)}</td>
                    <td className="p-3 text-right">{formatCurrency(c.custo_parada_h)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </section>

      {configOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-xl shadow-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold mb-4">Parâmetros financeiros</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Setor (vazio = padrão fábrica)</label>
                <select
                  value={configForm.sector_id}
                  onChange={(e) => setConfigForm((f) => ({ ...f, sector_id: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                >
                  <option value="">Padrão fábrica</option>
                  {sectors.map((s) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Valor de venda por peça OK (R$/un)</label>
                <input
                  type="number"
                  step="0.01"
                  value={configForm.valor_venda_ok}
                  onChange={(e) => setConfigForm((f) => ({ ...f, valor_venda_ok: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Custo por peça refugada (R$/un)</label>
                <input
                  type="number"
                  step="0.01"
                  value={configForm.custo_refugo_un}
                  onChange={(e) => setConfigForm((f) => ({ ...f, custo_refugo_un: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Custo de hora parada (R$/h)</label>
                <input
                  type="number"
                  step="0.01"
                  value={configForm.custo_parada_h}
                  onChange={(e) => setConfigForm((f) => ({ ...f, custo_parada_h: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button type="button" onClick={() => setConfigOpen(false)} className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">Cancelar</button>
              <button type="button" onClick={saveConfig} className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">Salvar</button>
            </div>
          </div>
        </div>
      )}

      {mappingOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-xl shadow-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold mb-4">Mapeamento de tags (CLP)</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Ativo / Linha</label>
                <select
                  value={mappingForm.asset_id}
                  onChange={(e) => setMappingForm((f) => ({ ...f, asset_id: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                >
                  <option value="">— Selecionar —</option>
                  {assets.map((a) => (
                    <option key={a.id} value={a.id}>{a.display_name || a.source_tag_id}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tag contador OK (peças boas)</label>
                <input
                  type="text"
                  value={mappingForm.tag_ok}
                  onChange={(e) => setMappingForm((f) => ({ ...f, tag_ok: e.target.value }))}
                  placeholder="ex: Total_Pecas"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tag contador NOK / refugo</label>
                <input
                  type="text"
                  value={mappingForm.tag_nok}
                  onChange={(e) => setMappingForm((f) => ({ ...f, tag_nok: e.target.value }))}
                  placeholder="ex: Refugo"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tag status (rodando/parado)</label>
                <input
                  type="text"
                  value={mappingForm.tag_status}
                  onChange={(e) => setMappingForm((f) => ({ ...f, tag_status: e.target.value }))}
                  placeholder="ex: Status_Producao"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Leitura do contador</label>
                <select
                  value={mappingForm.reading_rule}
                  onChange={(e) => setMappingForm((f) => ({ ...f, reading_rule: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                >
                  <option value="delta">Delta (variação no período)</option>
                  <option value="absolute">Valor absoluto</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button type="button" onClick={() => setMappingOpen(false)} className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">Fechar</button>
              <button type="button" onClick={saveMapping} className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">Salvar</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
