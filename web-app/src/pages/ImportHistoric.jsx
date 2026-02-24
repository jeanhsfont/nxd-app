import React, { useState, useEffect, useCallback } from 'react';
import {
  Download,
  Plus,
  RefreshCw,
  XCircle,
  RotateCcw,
  Send,
  FileJson,
  ChevronDown,
  ChevronRight,
  Loader2,
  CheckCircle2,
  AlertCircle,
  Clock,
} from 'lucide-react';
import api from '../utils/api';

const STATUS_LABELS = {
  pending: { label: 'Pendente', color: 'text-amber-600 bg-amber-50', Icon: Clock },
  running: { label: 'Em execução', color: 'text-blue-600 bg-blue-50', Icon: Loader2 },
  done: { label: 'Concluído', color: 'text-emerald-600 bg-emerald-50', Icon: CheckCircle2 },
  failed: { label: 'Falhou', color: 'text-red-600 bg-red-50', Icon: AlertCircle },
  cancelled: { label: 'Cancelado', color: 'text-gray-600 bg-gray-100', Icon: XCircle },
};

function JobRow({ job, assets, onRefresh, onCancel, onRetry, onSelectJob }) {
  const [expanded, setExpanded] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [jsonInput, setJsonInput] = useState('');
  const [submitError, setSubmitError] = useState('');
  const meta = STATUS_LABELS[job.status] || { label: job.status, color: 'text-gray-600', Icon: Clock };
  const Icon = meta.Icon;
  const canCancel = job.status === 'pending' || job.status === 'running';
  const canRetry = job.status === 'failed' || job.status === 'cancelled';
  const canSubmitData = job.status === 'pending' && job.source_type === 'memory';

  const handleCancel = async () => {
    try {
      await api.post(`/api/admin/import-jobs/${job.id}/cancel`);
      onRefresh();
    } catch (e) {
      console.error(e);
    }
  };

  const handleRetry = async () => {
    try {
      await api.post(`/api/admin/import-jobs/${job.id}/retry`);
      onRefresh();
    } catch (e) {
      console.error(e);
    }
  };

  const handleSubmitData = async () => {
    if (!jsonInput.trim()) {
      setSubmitError('Cole o JSON com o array de linhas.');
      return;
    }
    let rows;
    try {
      rows = JSON.parse(jsonInput);
      if (!Array.isArray(rows)) {
        setSubmitError('O JSON deve ser um array de objetos.');
        return;
      }
    } catch {
      setSubmitError('JSON inválido.');
      return;
    }
    const assetId = job.asset_id || (assets && assets[0]?.id);
    if (!assetId) {
      setSubmitError('Selecione um ativo no job ou cadastre um ativo antes.');
      return;
    }
    setSubmitError('');
    setSubmitting(true);
    try {
      await api.post(`/api/admin/import-jobs/${job.id}/data`, {
        asset_id: assetId,
        rows: rows.map((r) => ({
          ts: r.ts || r.timestamp,
          metric_key: r.metric_key || r.tag,
          metric_value: Number(r.metric_value ?? r.value ?? 0),
          status: r.status || 'OK',
        })),
      });
      setJsonInput('');
      onRefresh();
    } catch (e) {
      setSubmitError(e.response?.data?.error || e.message || 'Erro ao enviar dados.');
    } finally {
      setSubmitting(false);
    }
  };

  const progress =
    job.rows_total > 0
      ? Math.round((Number(job.rows_done || 0) / Number(job.rows_total)) * 100)
      : 0;

  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden bg-white">
      <div
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-3">
          {expanded ? (
            <ChevronDown className="w-5 h-5 text-gray-400" />
          ) : (
            <ChevronRight className="w-5 h-5 text-gray-400" />
          )}
          <span className="font-mono text-sm text-gray-500">{job.id?.slice(0, 8)}…</span>
          <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${meta.color}`}>
            {Icon === Loader2 && job.status === 'running' ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Icon className="w-3.5 h-3.5" />
            )}
            {meta.label}
          </span>
          {job.source_type && (
            <span className="text-xs text-gray-400">({job.source_type})</span>
          )}
        </div>
        <div className="flex items-center gap-4">
          {(job.rows_total > 0 || job.rows_done > 0) && (
            <span className="text-sm text-gray-500">
              {Number(job.rows_done || 0).toLocaleString()} / {Number(job.rows_total || 0).toLocaleString()}
              {job.status === 'running' && (
                <span className="ml-1 text-gray-400">({progress}%)</span>
              )}
            </span>
          )}
          <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
            {canCancel && (
              <button
                type="button"
                onClick={handleCancel}
                className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                title="Cancelar"
              >
                <XCircle className="w-4 h-4" />
              </button>
            )}
            {canRetry && (
              <button
                type="button"
                onClick={handleRetry}
                className="p-2 text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                title="Tentar novamente"
              >
                <RotateCcw className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>
      </div>

      {expanded && (
        <div className="border-t border-gray-200 bg-gray-50/80 p-4 space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm">
            <div>
              <span className="text-gray-500">Criado</span>
              <p className="font-medium text-gray-900">
                {job.created_at ? new Date(job.created_at).toLocaleString('pt-BR') : '—'}
              </p>
            </div>
            {job.finished_at && (
              <div>
                <span className="text-gray-500">Finalizado</span>
                <p className="font-medium text-gray-900">
                  {new Date(job.finished_at).toLocaleString('pt-BR')}
                </p>
              </div>
            )}
            {job.error_message && (
              <div className="col-span-2">
                <span className="text-gray-500">Erro</span>
                <p className="font-medium text-red-600 text-xs break-all">{job.error_message}</p>
              </div>
            )}
          </div>

          {canSubmitData && (
            <div className="border border-gray-200 rounded-lg p-4 bg-white">
              <h4 className="text-sm font-semibold text-gray-700 mb-2 flex items-center gap-2">
                <FileJson className="w-4 h-4" />
                Enviar dados (JSON)
              </h4>
              <p className="text-xs text-gray-500 mb-2">
                Array de objetos com: ts (RFC3339), metric_key, metric_value, status (opcional).
              </p>
              <textarea
                value={jsonInput}
                onChange={(e) => setJsonInput(e.target.value)}
                placeholder='[{"ts":"2024-01-15T10:00:00Z","metric_key":"Temperatura","metric_value":85.2}, ...]'
                className="w-full h-32 font-mono text-sm border border-gray-300 rounded-lg p-3 resize-y"
              />
              {submitError && (
                <p className="text-sm text-red-600 mt-1">{submitError}</p>
              )}
              <button
                type="button"
                disabled={submitting}
                onClick={handleSubmitData}
                className="mt-2 inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              >
                {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                Enviar dados
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function ImportHistoric() {
  const [jobs, setJobs] = useState([]);
  const [assets, setAssets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [createForm, setCreateForm] = useState({
    source_type: 'memory',
    asset_id: '',
    dx_url: '',
    batch_size: 1000,
  });

  const fetchJobs = useCallback(async () => {
    try {
      const [jobsRes, dashRes] = await Promise.all([
        api.get('/api/admin/import-jobs'),
        api.get('/api/dashboard/data').catch(() => ({ data: {} })),
      ]);
      setJobs(jobsRes.data?.jobs || []);
      setAssets(dashRes.data?.assets || []);
    } catch (e) {
      console.error(e);
      setJobs([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchJobs();
    const t = setInterval(fetchJobs, 8000);
    return () => clearInterval(t);
  }, [fetchJobs]);

  const handleCreate = async () => {
    if (createForm.source_type === 'dx_http') {
      if (!createForm.dx_url?.trim()) return;
      if (!createForm.asset_id) return;
    }
    setCreating(true);
    try {
      const payload = {
        source_type: createForm.source_type,
        batch_size: createForm.batch_size || 1000,
      };
      if (createForm.asset_id) payload.asset_id = createForm.asset_id;
      if (createForm.source_type === 'dx_http') {
        payload.source_config = { url: createForm.dx_url.trim(), asset_id: createForm.asset_id };
      }
      await api.post('/api/admin/import-jobs', payload);
      setCreateOpen(false);
      setCreateForm({ source_type: 'memory', asset_id: '', dx_url: '', batch_size: 1000 });
      fetchJobs();
    } catch (e) {
      console.error(e);
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="p-8 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Download className="w-8 h-8 text-indigo-600" />
            Importar histórico
          </h1>
          <p className="text-gray-500 mt-1">
            Crie jobs para carregar dados históricos de telemetria (Download Longo). Envie JSON ou use um script que chame a API.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={fetchJobs}
            className="p-2 text-gray-500 hover:bg-gray-100 rounded-lg"
            title="Atualizar"
          >
            <RefreshCw className="w-5 h-5" />
          </button>
          <button
            type="button"
            onClick={() => setCreateOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
          >
            <Plus className="w-5 h-5" />
            Novo job
          </button>
        </div>
      </div>

      {createOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-xl shadow-xl p-6 w-full max-w-md">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Novo job de importação</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tipo</label>
                <select
                  value={createForm.source_type}
                  onChange={(e) => setCreateForm((f) => ({ ...f, source_type: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                >
                  <option value="memory">Envio manual (memory)</option>
                  <option value="dx_http">DX/CLP via HTTP (puxar histórico do DX)</option>
                </select>
              </div>
              {createForm.source_type === 'dx_http' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">URL do DX (GET)</label>
                    <input
                      type="url"
                      value={createForm.dx_url || ''}
                      onChange={(e) => setCreateForm((f) => ({ ...f, dx_url: e.target.value }))}
                      placeholder="https://dx.empresa.com/api/history"
                      className="w-full border border-gray-300 rounded-lg px-3 py-2"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Ativo (obrigatório para dx_http)</label>
                    <select
                      value={createForm.asset_id}
                      onChange={(e) => setCreateForm((f) => ({ ...f, asset_id: e.target.value }))}
                      className="w-full border border-gray-300 rounded-lg px-3 py-2"
                    >
                      <option value="">— Selecionar —</option>
                      {assets.map((a) => (
                        <option key={a.id} value={a.id}>{a.display_name || a.source_tag_id}</option>
                      ))}
                    </select>
                  </div>
                </>
              )}
              {createForm.source_type === 'memory' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Ativo (opcional)</label>
                  <select
                    value={createForm.asset_id}
                    onChange={(e) => setCreateForm((f) => ({ ...f, asset_id: e.target.value }))}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2"
                  >
                    <option value="">— Selecionar —</option>
                    {assets.map((a) => (
                      <option key={a.id} value={a.id}>{a.display_name || a.source_tag_id}</option>
                    ))}
                  </select>
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tamanho do lote</label>
                <input
                  type="number"
                  min={100}
                  max={10000}
                  value={createForm.batch_size}
                  onChange={(e) =>
                    setCreateForm((f) => ({ ...f, batch_size: parseInt(e.target.value, 10) || 1000 }))
                  }
                  className="w-full border border-gray-300 rounded-lg px-3 py-2"
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button
                type="button"
                onClick={() => setCreateOpen(false)}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={handleCreate}
                disabled={creating}
                className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 flex items-center gap-2"
              >
                {creating && <Loader2 className="w-4 h-4 animate-spin" />}
                Criar job
              </button>
            </div>
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="w-8 h-8 animate-spin text-indigo-600" />
        </div>
      ) : jobs.length === 0 ? (
        <div className="text-center py-16 bg-gray-50 rounded-xl border border-gray-200">
          <Download className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <p className="text-gray-500">Nenhum job de importação ainda.</p>
          <p className="text-sm text-gray-400 mt-1">Clique em &quot;Novo job&quot; para criar um e enviar dados históricos.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {jobs.map((job) => (
            <JobRow
              key={job.id}
              job={job}
              assets={assets}
              onRefresh={fetchJobs}
              onCancel={() => {}}
              onRetry={() => {}}
              onSelectJob={() => {}}
            />
          ))}
        </div>
      )}
    </div>
  );
}
