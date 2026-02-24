import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { Activity, WifiOff, RefreshCw, Zap, Thermometer, BarChart2, Clock, Box, AlertCircle } from 'lucide-react';

const METRIC_ICONS = {
  Temperatura_Molde: Thermometer,
  Consumo_Energia_kWh: Zap,
  Health_Score: Activity,
  Total_Pecas: Box,
  Status_Producao: Activity,
  Custo_Hora_Parada: BarChart2,
  Cycle_Time_ms: Clock,
  Fault_Code: AlertCircle,
};

const METRIC_UNITS = {
  Temperatura_Molde: '°C',
  Consumo_Energia_kWh: 'kWh',
  Health_Score: '%',
  Total_Pecas: 'pç',
  Custo_Hora_Parada: 'R$/h',
  Cycle_Time_ms: 'ms',
  Fault_Code: '',
  Status_Producao: '',
};

function formatValue(key, value) {
  if (key === 'Health_Score') return (value * 100).toFixed(1);
  if (key === 'Status_Producao') return value ? 'ON' : 'OFF';
  if (typeof value === 'number') return value % 1 === 0 ? value.toString() : value.toFixed(2);
  return value;
}

function formatRelativeTime(dateStr) {
  if (!dateStr) return 'nunca';
  const diff = (Date.now() - new Date(dateStr)) / 1000;
  if (diff < 5) return 'agora';
  if (diff < 60) return `${Math.floor(diff)}s atrás`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m atrás`;
  return `${Math.floor(diff / 3600)}h atrás`;
}

function PulseRing({ online }) {
  return (
    <span className="relative flex h-3 w-3">
      {online && (
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
      )}
      <span className={`relative inline-flex rounded-full h-3 w-3 ${online ? 'bg-emerald-500' : 'bg-slate-600'}`} />
    </span>
  );
}

function MetricPill({ metricKey, value }) {
  const Icon = METRIC_ICONS[metricKey] || BarChart2;
  const unit = METRIC_UNITS[metricKey] ?? '';
  const formatted = formatValue(metricKey, value);
  const label = metricKey.replace(/_/g, ' ');
  const isAlert = metricKey === 'Fault_Code' && value > 0;
  const isGood = metricKey === 'Health_Score' && value > 0.9;

  return (
    <div className={`flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-mono transition-all
      ${isAlert ? 'bg-red-500/10 border border-red-500/30 text-red-300' :
        isGood ? 'bg-emerald-500/10 border border-emerald-500/30 text-emerald-300' :
        'bg-white/5 border border-white/10 text-slate-300'}`}>
      <Icon className="w-3.5 h-3.5 shrink-0 opacity-60" />
      <span className="text-slate-400 truncate max-w-[90px]">{label}</span>
      <span className={`ml-auto font-bold tabular-nums ${isAlert ? 'text-red-400' : isGood ? 'text-emerald-400' : 'text-white'}`}>
        {formatted}{unit && <span className="text-slate-500 font-normal ml-0.5">{unit}</span>}
      </span>
    </div>
  );
}

function AssetCard({ asset, index }) {
  const [expanded, setExpanded] = useState(false);
  const metrics = asset.metrics || {};
  const metricEntries = Object.entries(metrics);
  const visibleMetrics = expanded ? metricEntries : metricEntries.slice(0, 4);
  const healthScore = metrics['Health_Score'];
  const healthPct = healthScore != null ? Math.round(healthScore * 100) : null;
  const isProducing = metrics['Status_Producao'];

  return (
    <div
      className="group relative bg-[#0f1117] border border-white/[0.07] rounded-2xl overflow-hidden transition-all duration-300 hover:border-white/20 hover:shadow-2xl hover:shadow-black/50"
      style={{ animationDelay: `${index * 60}ms` }}
    >
      <div className={`h-0.5 w-full ${asset.is_online ? 'bg-gradient-to-r from-emerald-500 via-teal-400 to-transparent' : 'bg-gradient-to-r from-slate-700 to-transparent'}`} />
      <div className="p-5">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <PulseRing online={asset.is_online} />
              <span className={`text-xs font-semibold uppercase tracking-widest ${asset.is_online ? 'text-emerald-400' : 'text-slate-500'}`}>
                {asset.is_online ? 'Online' : 'Offline'}
              </span>
            </div>
            <h3 className="text-white font-bold text-base leading-tight truncate">{asset.display_name}</h3>
            <p className="text-slate-500 text-xs mt-0.5 font-mono">{asset.source_tag_id}</p>
          </div>
          {healthPct != null && (
            <div className="relative w-12 h-12 shrink-0 ml-3">
              <svg viewBox="0 0 36 36" className="w-12 h-12 -rotate-90">
                <circle cx="18" cy="18" r="15.9" fill="none" stroke="#1e2533" strokeWidth="3" />
                <circle
                  cx="18" cy="18" r="15.9" fill="none"
                  stroke={healthPct > 90 ? '#10b981' : healthPct > 70 ? '#f59e0b' : '#ef4444'}
                  strokeWidth="3"
                  strokeDasharray={`${healthPct} 100`}
                  strokeLinecap="round"
                  className="transition-all duration-700"
                />
              </svg>
              <span className="absolute inset-0 flex items-center justify-center text-[10px] font-bold text-white">{healthPct}%</span>
            </div>
          )}
        </div>

        {metricEntries.length > 0 ? (
          <>
            <div className="grid grid-cols-2 gap-1.5 mb-2">
              {visibleMetrics.map(([key, val]) => (
                <MetricPill key={key} metricKey={key} value={val} />
              ))}
            </div>
            {metricEntries.length > 4 && (
              <button
                onClick={() => setExpanded(!expanded)}
                className="text-xs text-slate-500 hover:text-slate-300 transition-colors mt-1 w-full text-center py-1"
              >
                {expanded ? '▲ menos' : `▼ +${metricEntries.length - 4} métricas`}
              </button>
            )}
          </>
        ) : (
          <div className="text-slate-600 text-xs text-center py-4 border border-dashed border-white/5 rounded-lg">
            aguardando dados...
          </div>
        )}

        <div className="flex items-center justify-between mt-3 pt-3 border-t border-white/[0.06]">
          <div className="flex items-center gap-1.5 text-slate-600 text-xs">
            <Clock className="w-3 h-3" />
            <span>{formatRelativeTime(asset.last_seen)}</span>
          </div>
          {isProducing !== undefined && (
            <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${isProducing ? 'bg-emerald-500/15 text-emerald-400' : 'bg-slate-700/50 text-slate-400'}`}>
              {isProducing ? '⚡ Produzindo' : '⏸ Parado'}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, sub, accent }) {
  return (
    <div className="bg-[#0f1117] border border-white/[0.07] rounded-2xl p-5">
      <p className="text-xs text-slate-500 uppercase tracking-widest font-semibold mb-2">{label}</p>
      <p className={`text-4xl font-black tabular-nums ${accent}`}>{value}</p>
      {sub && <p className="text-xs text-slate-600 mt-1">{sub}</p>}
    </div>
  );
}

// ─── Loading overlay real (sem progresso simulado) ────────────────────────────
function DashboardLoadingOverlay({ slow }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#080b10]/90 backdrop-blur-sm" aria-live="polite" aria-busy="true">
      <div className="bg-[#0f1117] border border-white/10 rounded-2xl p-8 w-full max-w-sm mx-4 shadow-2xl shadow-black/60 text-center">
        <div className="w-12 h-12 rounded-xl bg-teal-500/10 border border-teal-500/20 flex items-center justify-center mx-auto mb-4">
          <Activity className="w-6 h-6 text-teal-400 animate-pulse" aria-hidden="true" />
        </div>
        <p className="text-white font-semibold text-sm mb-1">Carregando telemetria…</p>
        <p className="text-slate-500 text-xs">
          {slow ? 'Aguardando resposta do servidor. Verifique a conexão.' : 'Aguarde os dados em tempo real.'}
        </p>
      </div>
    </div>
  );
}

export default function Dashboard() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [loadingSlow, setLoadingSlow] = useState(false);
  const [error, setError] = useState('');
  const [lastRefresh, setLastRefresh] = useState(null);
  const [spinning, setSpinning] = useState(false);
  const intervalRef = useRef(null);
  const slowTimerRef = useRef(null);

  useEffect(() => {
    if (!loading) {
      setLoadingSlow(false);
      if (slowTimerRef.current) clearTimeout(slowTimerRef.current);
      return;
    }
    slowTimerRef.current = setTimeout(() => setLoadingSlow(true), 5000);
    return () => {
      if (slowTimerRef.current) clearTimeout(slowTimerRef.current);
    };
  }, [loading]);

  // Guard: never start a new fetch if the previous one is still in flight
  const fetchingRef = useRef(false);

  const fetchData = useCallback(async (manual = false) => {
    if (fetchingRef.current) return; // prevent overlapping requests
    fetchingRef.current = true;
    if (manual) setSpinning(true);
    try {
      const token = localStorage.getItem('nxd-token');
      const res = await fetch('/api/dashboard/data', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error(`Erro ${res.status}`);
      const json = await res.json();
      setData(json);
      setLastRefresh(new Date());
      setError('');
    } catch (e) {
      setError(e.message);
    } finally {
      fetchingRef.current = false;
      setLoading(false);
      if (manual) setTimeout(() => setSpinning(false), 600);
    }
  }, []);

  useEffect(() => {
    fetchData();
    intervalRef.current = setInterval(() => fetchData(), 10000);
    return () => clearInterval(intervalRef.current);
  }, [fetchData]);

  const onlineAssets = data?.online_assets ?? 0;
  const totalAssets = data?.total_assets ?? 0;
  const offlineAssets = totalAssets - onlineAssets;
  const assets = data?.assets ?? [];

  return (
    <div className="min-h-screen bg-[#080b10] text-white">
      {/* First-load modal: shown only while the very first fetch is in flight */}
      {loading && <DashboardLoadingOverlay slow={loadingSlow} />}

      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute -top-40 -left-40 w-96 h-96 bg-teal-500/5 rounded-full blur-3xl" />
        <div className="absolute top-1/2 -right-40 w-80 h-80 bg-indigo-500/5 rounded-full blur-3xl" />
      </div>

      <div className="relative z-10 max-w-7xl mx-auto px-6 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <div className="w-8 h-8 rounded-lg bg-teal-500/10 border border-teal-500/20 flex items-center justify-center">
                <Activity className="w-4 h-4 text-teal-400" />
              </div>
              <h1 className="text-2xl font-black tracking-tight">
                {data?.factory_name || 'Dashboard'}
              </h1>
            </div>
            <p className="text-slate-500 text-sm ml-11">
              {lastRefresh ? `Atualizado ${formatRelativeTime(lastRefresh)}` : 'Carregando...'}
              <span className="mx-2 text-slate-700">·</span>
              <span className="text-slate-600">Auto-refresh 10s</span>
            </p>
          </div>

          <button
            onClick={() => fetchData(true)}
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl text-sm text-slate-300 transition-all disabled:opacity-40 disabled:cursor-not-allowed"
          >
            <RefreshCw className={`w-4 h-4 ${spinning ? 'animate-spin' : ''}`} />
            Atualizar
          </button>
        </div>

        {error && (
          <div className="mb-6 px-4 py-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            {error}
          </div>
        )}

        {/* Erro na carga inicial: mensagem clara + CTA (não fica em limbo) */}
        {error && !data && (
          <div className="mb-8 p-6 bg-[#0f1117] border border-red-500/20 rounded-2xl text-center">
            <p className="text-red-400 font-medium mb-2">Não foi possível carregar os dados.</p>
            <p className="text-slate-500 text-sm mb-4">{error}</p>
            <button
              type="button"
              onClick={() => { setError(''); setLoading(true); fetchData(true); }}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <RefreshCw className="w-4 h-4" />
              Tentar novamente
            </button>
          </div>
        )}

        {/* Stat cards — ocultos quando erro na carga inicial (só mostramos CTA) */}
        {!(error && !data) && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <StatCard label="Total de CLPs" value={loading ? '—' : totalAssets} sub="ativos detectados" accent="text-white" />
          <StatCard label="Online" value={loading ? '—' : onlineAssets} sub="últimos 60s" accent="text-emerald-400" />
          <StatCard label="Offline" value={loading ? '—' : offlineAssets} sub="sem sinal recente" accent={offlineAssets > 0 ? 'text-amber-400' : 'text-slate-500'} />
          <StatCard
            label="Disponibilidade"
            value={loading || totalAssets === 0 ? '—' : `${Math.round((onlineAssets / totalAssets) * 100)}%`}
            sub="online / total"
            accent={onlineAssets === totalAssets && totalAssets > 0 ? 'text-emerald-400' : 'text-amber-400'}
          />
        </div>

        {/* Connection bar */}
        {totalAssets > 0 && (
          <div className="mb-8">
            <div className="flex items-center justify-between text-xs text-slate-500 mb-2">
              <span className="uppercase tracking-wider font-semibold">Status de conexão</span>
              <span>{onlineAssets}/{totalAssets} CLPs</span>
            </div>
            <div className="h-2 bg-white/5 rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-emerald-500 to-teal-400 rounded-full transition-all duration-700"
                style={{ width: totalAssets > 0 ? `${(onlineAssets / totalAssets) * 100}%` : '0%' }}
              />
            </div>
          </div>
        )}

        {/* Assets grid – skeleton only shown if loading AND no data yet */}
        {loading && !data ? (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="bg-[#0f1117] border border-white/[0.07] rounded-2xl h-52 animate-pulse" />
            ))}
          </div>
        ) : assets.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24 text-center max-w-md mx-auto">
            <div className="w-16 h-16 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mb-4">
              <WifiOff className="w-7 h-7 text-slate-600" aria-hidden="true" />
            </div>
            <p className="text-slate-400 font-semibold mb-2">Nenhum CLP detectado ainda</p>
            <p className="text-slate-600 text-sm mb-6">
              Siga os passos abaixo para conectar seu DX e receber telemetria em tempo real.
            </p>
            <ol className="text-left text-sm text-slate-400 space-y-3 mb-6 w-full">
              <li className="flex items-center gap-3">
                <span className="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-300 font-bold text-xs flex items-center justify-center">1</span>
                Vá em <strong className="text-slate-300">Ajustes</strong> e copie sua API Key
              </li>
              <li className="flex items-center gap-3">
                <span className="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-300 font-bold text-xs flex items-center justify-center">2</span>
                Configure a API Key no DX Simulator ou no seu dispositivo
              </li>
              <li className="flex items-center gap-3">
                <span className="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-300 font-bold text-xs flex items-center justify-center">3</span>
                Inicie o envio de dados (ex.: <code className="bg-white/5 px-1.5 py-0.5 rounded text-xs">dx-simulator</code>)
              </li>
              <li className="flex items-center gap-3">
                <span className="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-300 font-bold text-xs flex items-center justify-center">4</span>
                Volte aqui para ver os CLPs e métricas
              </li>
            </ol>
            <Link
              to="/settings"
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-[#080b10]"
            >
              Ir para Ajustes
            </Link>
            <div className="mt-6 px-4 py-2 bg-white/5 border border-white/10 rounded-xl font-mono text-xs text-slate-500">
              cd dx-simulator &amp;&amp; go run main.go
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {assets.map((asset, i) => (
              <AssetCard key={asset.id} asset={asset} index={i} />
            ))}
          </div>
        )}

        {assets.length > 0 && (
          <div className="mt-8 flex items-center justify-center gap-2 text-xs text-slate-600">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-teal-400 opacity-40" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-teal-500" />
            </span>
            Telemetria ao vivo · {assets.filter(a => a.is_online).map(a => a.source_tag_id).join(', ') || 'nenhum online'}
          </div>
        )}
        )}
      </div>
    </div>
  );
}
