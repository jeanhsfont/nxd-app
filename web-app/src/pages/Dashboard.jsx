import React, { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { Activity, RefreshCw, Thermometer, Zap, BarChart2, Clock, Box, AlertCircle, Factory } from 'lucide-react';
import api from '../utils/api';

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

function MetricPill({ metricKey, value }) {
  const Icon = METRIC_ICONS[metricKey] || BarChart2;
  const unit = METRIC_UNITS[metricKey] ?? '';
  const formatted = formatValue(metricKey, value);
  const label = metricKey.replace(/_/g, ' ');
  const isAlert = metricKey === 'Fault_Code' && value > 0;
  const isGood = metricKey === 'Health_Score' && value > 0.9;

  return (
    <div className={`flex items-center gap-2 px-3 py-2 rounded-lg text-xs transition-all ${
      isAlert ? 'nxd-badge-danger' : isGood ? 'nxd-badge-success' : 'nxd-badge-gray'
    }`}>
      <Icon className="w-3.5 h-3.5 shrink-0" />
      <span className="truncate max-w-[90px]">{label}</span>
      <span className="ml-auto font-bold tabular-nums">
        {formatted}{unit && <span className="opacity-60 font-normal ml-0.5">{unit}</span>}
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
    <div className="nxd-card fade-in" style={{ animationDelay: `${index * 60}ms` }}>
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <span className={`status-dot ${asset.is_online ? 'status-online' : 'status-offline'}`}></span>
            <span className={`text-xs font-semibold uppercase tracking-wide ${
              asset.is_online ? 'text-green' : 'text-gray'
            }`}>
              {asset.is_online ? 'Online' : 'Offline'}
            </span>
          </div>
          <h3 className="text-gray-900 font-bold text-lg leading-tight truncate">{asset.display_name}</h3>
          <p className="text-gray-500 text-sm mt-1">{asset.source_tag_id}</p>
        </div>
        {healthPct != null && (
          <div className="relative w-14 h-14 shrink-0 ml-3">
            <svg viewBox="0 0 36 36" className="w-14 h-14 -rotate-90">
              <circle cx="18" cy="18" r="15.9" fill="none" stroke="var(--nxd-gray-200)" strokeWidth="3" />
              <circle
                cx="18" cy="18" r="15.9" fill="none"
                stroke={healthPct > 90 ? 'var(--nxd-green)' : healthPct > 70 ? '#f59e0b' : 'var(--nxd-red)'}
                strokeWidth="3"
                strokeDasharray={`${healthPct} 100`}
                strokeLinecap="round"
                className="transition-all duration-700"
              />
            </svg>
            <span className="absolute inset-0 flex items-center justify-center text-xs font-bold text-gray-900">{healthPct}%</span>
          </div>
        )}
      </div>

      {metricEntries.length > 0 ? (
        <>
          <div className="grid grid-cols-2 gap-2 mb-3">
            {visibleMetrics.map(([key, val]) => (
              <MetricPill key={key} metricKey={key} value={val} />
            ))}
          </div>
          {metricEntries.length > 4 && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="text-xs text-gray-500 hover:text-navy transition-colors w-full text-center py-2 border-t border-gray-200"
            >
              {expanded ? '▲ Mostrar menos' : `▼ Ver mais ${metricEntries.length - 4} métricas`}
            </button>
          )}
        </>
      ) : (
        <div className="text-gray-400 text-sm text-center py-6 border-2 border-dashed border-gray-200 rounded-lg">
          Aguardando dados...
        </div>
      )}

      <div className="flex items-center justify-between mt-4 pt-4 border-t border-gray-200">
        <div className="flex items-center gap-2 text-gray-500 text-xs">
          <Clock className="w-4 h-4" />
          <span>{formatRelativeTime(asset.last_seen)}</span>
        </div>
        {isProducing !== undefined && (
          <span className={`text-xs px-3 py-1 rounded-full font-medium ${
            isProducing ? 'nxd-badge-success' : 'nxd-badge-gray'
          }`}>
            {isProducing ? '⚡ Produzindo' : '⏸ Parado'}
          </span>
        )}
      </div>
    </div>
  );
}

function StatCard({ label, value, sub, isGreen }) {
  return (
    <div className="nxd-card">
      <p className="text-sm text-gray-500 uppercase tracking-wide font-semibold mb-2">{label}</p>
      <p className={`text-4xl font-black tabular-nums ${isGreen ? 'text-green' : 'text-navy'}`}>{value}</p>
      {sub && <p className="text-xs text-gray-500 mt-2">{sub}</p>}
    </div>
  );
}

export default function Dashboard() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [lastRefresh, setLastRefresh] = useState(null);
  const [spinning, setSpinning] = useState(false);
  const intervalRef = useRef(null);

  const fetchData = async () => {
    try {
      const response = await api.get('/api/dashboard/data');
      setData(response.data);
      setError('');
      setLastRefresh(new Date());
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    intervalRef.current = setInterval(fetchData, 5000);
    return () => clearInterval(intervalRef.current);
  }, []);

  const handleRefresh = () => {
    setSpinning(true);
    fetchData();
    setTimeout(() => setSpinning(false), 600);
  };

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="spinner mb-4"></div>
          <p className="text-gray-600">Carregando dashboard...</p>
        </div>
      </div>
    );
  }

  if (error && !data) {
    return (
      <div className="h-screen flex items-center justify-center bg-gray-50">
        <div className="nxd-card max-w-md text-center">
          <AlertCircle className="w-12 h-12 text-red mx-auto mb-4" />
          <h2 className="text-xl font-bold text-gray-900 mb-2">Erro ao Carregar</h2>
          <p className="text-gray-600 mb-4">{error}</p>
          <button onClick={handleRefresh} className="nxd-btn nxd-btn-primary">
            <RefreshCw className="w-4 h-4" />
            Tentar Novamente
          </button>
        </div>
      </div>
    );
  }

  const assets = data?.assets || [];
  const onlineCount = assets.filter(a => a.is_online).length;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto p-6">
        {/* Header */}
        <div className="page-header">
          <div className="page-header-icon">
            <Factory className="w-6 h-6" />
          </div>
          <div className="flex-1">
            <h1 className="page-title">Dashboard</h1>
            <p className="page-subtitle">Visão geral da sua operação em tempo real</p>
          </div>
          <button
            onClick={handleRefresh}
            className="nxd-btn nxd-btn-primary"
            disabled={spinning}
          >
            <RefreshCw className={`w-4 h-4 ${spinning ? 'animate-spin' : ''}`} />
            Atualizar
          </button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <StatCard label="Total de Ativos" value={data?.total_assets || 0} />
          <StatCard label="Ativos Online" value={onlineCount} isGreen sub={`${assets.length > 0 ? Math.round((onlineCount/assets.length)*100) : 0}% do total`} />
          <StatCard label="Última Atualização" value={lastRefresh ? lastRefresh.toLocaleTimeString('pt-BR') : '--:--'} />
        </div>

        {/* Assets Grid */}
        {assets.length === 0 ? (
          <div className="nxd-card text-center py-12">
            <Factory className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-xl font-bold text-gray-900 mb-2">Nenhum ativo encontrado</h3>
            <p className="text-gray-600 mb-6">Configure seus ativos para começar a monitorar</p>
            <Link to="/assets">
              <button className="nxd-btn nxd-btn-primary">
                Gerenciar Ativos
              </button>
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
            {assets.map((asset, idx) => (
              <AssetCard key={asset.source_tag_id} asset={asset} index={idx} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
