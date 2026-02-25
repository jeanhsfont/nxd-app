import React, { useState, useEffect } from 'react';
import { Settings as SettingsIcon, Key, Shield, CheckCircle, X, Loader2 } from 'lucide-react';
import api from '../utils/api';
import RegenerateKeyModal from '../components/RegenerateKeyModal';

export default function Settings() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [twoFaEnabled, setTwoFaEnabled] = useState(false);
  const [twoFaLoading, setTwoFaLoading] = useState(true);
  const [showSetup, setShowSetup] = useState(false);
  const [qrCodeUrl, setQrCodeUrl] = useState(null);
  const [confirmCode, setConfirmCode] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [showDisableInput, setShowDisableInput] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [message, setMessage] = useState(null);

  useEffect(() => {
    fetchTwoFaStatus();
  }, []);

  const fetchTwoFaStatus = async () => {
    setTwoFaLoading(true);
    try {
      const res = await api.get('/api/auth/2fa/status');
      setTwoFaEnabled(res.data.enabled);
    } catch {
      setTwoFaEnabled(false);
    } finally {
      setTwoFaLoading(false);
    }
  };

  const handleSetup2FA = async () => {
    setActionLoading(true);
    setMessage(null);
    try {
      const res = await api.get('/api/auth/2fa/setup', { responseType: 'blob' });
      const url = URL.createObjectURL(res.data);
      setQrCodeUrl(url);
      setShowSetup(true);
      setConfirmCode('');
    } catch {
      setMessage({ type: 'error', text: 'Erro ao gerar QR code. Tente novamente.' });
    } finally {
      setActionLoading(false);
    }
  };

  const handleConfirm2FA = async () => {
    if (confirmCode.length !== 6) {
      setMessage({ type: 'error', text: 'O código deve ter 6 dígitos.' });
      return;
    }
    setActionLoading(true);
    setMessage(null);
    try {
      await api.post('/api/auth/2fa/confirm', { code: confirmCode });
      setTwoFaEnabled(true);
      setShowSetup(false);
      setQrCodeUrl(null);
      setConfirmCode('');
      setMessage({ type: 'success', text: '2FA ativado com sucesso! Sua conta está protegida.' });
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data || 'Código inválido. Tente novamente.' });
    } finally {
      setActionLoading(false);
    }
  };

  const handleDisable2FA = async () => {
    if (!disableCode) {
      setMessage({ type: 'error', text: 'Informe o código do app autenticador.' });
      return;
    }
    setActionLoading(true);
    setMessage(null);
    try {
      await api.post('/api/auth/2fa/disable', { code: disableCode });
      setTwoFaEnabled(false);
      setShowDisableInput(false);
      setDisableCode('');
      setMessage({ type: 'success', text: '2FA desativado.' });
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data || 'Código inválido.' });
    } finally {
      setActionLoading(false);
    }
  };

  const cancelSetup = () => {
    setShowSetup(false);
    setQrCodeUrl(null);
    setConfirmCode('');
    setMessage(null);
  };

  return (
    <>
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-4xl mx-auto p-6">
          <div className="page-header">
            <div className="page-header-icon">
              <SettingsIcon className="w-6 h-6" />
            </div>
            <div>
              <h1 className="page-title">Configurações</h1>
              <p className="page-subtitle">Gerencie sua chave de API e segurança da conta</p>
            </div>
          </div>

          {message && (
            <div className={`fade-in mb-6 flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium border ${
              message.type === 'success'
                ? 'bg-green/10 text-green border-green/20'
                : 'bg-red/10 text-red border-red/20'
            }`}>
              {message.type === 'success' ? <CheckCircle className="w-5 h-5" /> : <X className="w-5 h-5" />}
              {message.text}
            </div>
          )}

          {/* API Key Card */}
          <div className="nxd-card mb-6">
            <div className="flex items-center gap-3 mb-5 pb-5 border-b border-gray-200">
              <div className="w-10 h-10 rounded-lg bg-navy/10 flex items-center justify-center text-navy">
                <Key className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-900">Chave de API</h2>
                <p className="text-xs text-gray-500">Para conectar o DX Simulator e dispositivos externos</p>
              </div>
            </div>
            <p className="text-sm text-gray-600 mb-5 leading-relaxed">
              Ao gerar uma nova chave, sua chave atual será <span className="font-medium text-gray-900">invalidada permanentemente</span>.
              A nova chave só será exibida uma única vez por segurança.
            </p>
            <button
              onClick={() => setIsModalOpen(true)}
              className="inline-flex items-center gap-2 bg-red/10 hover:bg-red/20 text-red font-semibold text-sm py-2.5 px-4 rounded-lg border border-red/20 transition-colors"
            >
              <Key className="w-4 h-4" />
              Gerar Nova Chave de API
            </button>
          </div>

          {/* 2FA Card */}
          <div className="nxd-card">
            <div className="flex items-center justify-between mb-5 pb-5 border-b border-gray-200">
              <div className="flex items-center gap-3">
                <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                  twoFaEnabled ? 'bg-green/10 text-green' : 'bg-gray-200 text-gray-500'
                }`}>
                  <Shield className="w-5 h-5" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">Autenticação de Dois Fatores</h2>
                  <p className="text-xs text-gray-500">TOTP — Google Authenticator, Authy, etc.</p>
                </div>
              </div>
              {!twoFaLoading && (
                <span className={`nxd-badge ${twoFaEnabled ? 'nxd-badge-success' : 'nxd-badge-gray'}`}>
                  {twoFaEnabled ? <><CheckCircle className="w-3 h-3" /> Ativo</> : 'Inativo'}
                </span>
              )}
            </div>

            {twoFaLoading ? (
              <div className="flex items-center gap-2 text-gray-400 text-sm py-2">
                <Loader2 className="w-4 h-4 animate-spin" /> Verificando status...
              </div>
            ) : twoFaEnabled ? (
              <div className="fade-in">
                <div className="flex items-start gap-3 p-4 bg-green/10 rounded-lg border border-green/20 mb-5">
                  <CheckCircle className="w-5 h-5 text-green flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="text-sm font-semibold text-gray-900">Conta protegida com 2FA</p>
                    <p className="text-xs text-gray-600 mt-1">Cada acesso exigirá confirmação pelo app autenticador.</p>
                  </div>
                </div>

                {!showDisableInput ? (
                  <button
                    onClick={() => { setShowDisableInput(true); setMessage(null); }}
                    className="inline-flex items-center gap-2 text-sm text-gray-500 hover:text-red font-medium py-2 px-3 rounded-lg hover:bg-red/10 transition-colors"
                  >
                    <X className="w-4 h-4" /> Desativar 2FA
                  </button>
                ) : (
                  <div className="fade-in space-y-3">
                    <p className="text-sm text-gray-700 font-medium">Confirme com o código do seu app autenticador:</p>
                    <div className="flex gap-3">
                      <input
                        type="text"
                        inputMode="numeric"
                        maxLength={6}
                        value={disableCode}
                        onChange={e => setDisableCode(e.target.value.replace(/\D/g, ''))}
                        placeholder="000000"
                        className="nxd-input w-36 text-center text-lg font-mono tracking-widest"
                      />
                      <button
                        onClick={handleDisable2FA}
                        disabled={actionLoading}
                        className="nxd-btn bg-red hover:bg-red text-white"
                      >
                        {actionLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <X className="w-4 h-4" />}
                        Desativar
                      </button>
                      <button
                        onClick={() => { setShowDisableInput(false); setDisableCode(''); setMessage(null); }}
                        className="text-sm text-gray-500 hover:text-gray-700 px-3 py-2 rounded-lg hover:bg-gray-100 transition-colors"
                      >
                        Cancelar
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ) : !showSetup ? (
              <div className="fade-in">
                <p className="text-sm text-gray-600 mb-5 leading-relaxed">
                  Adicione uma camada extra de segurança à sua conta. Após ativar, você precisará de um
                  código temporário do app autenticador para fazer login.
                </p>
                <button
                  onClick={handleSetup2FA}
                  disabled={actionLoading}
                  className="nxd-btn nxd-btn-primary"
                >
                  {actionLoading ? <Loader2 className="w-5 h-5 animate-spin" /> : <Shield className="w-5 h-5" />}
                  {actionLoading ? 'Gerando QR Code...' : 'Ativar 2FA'}
                </button>
              </div>
            ) : (
              <div className="fade-in space-y-6">
                <div className="flex items-start gap-4">
                  <div className="flex-shrink-0">
                    {qrCodeUrl ? (
                      <div className="w-44 h-44 rounded-lg border-2 border-navy/20 bg-white p-2">
                        <img src={qrCodeUrl} alt="QR Code 2FA" className="w-full h-full rounded" />
                      </div>
                    ) : (
                      <div className="w-44 h-44 rounded-lg border-2 border-dashed border-gray-200 bg-gray-50 flex items-center justify-center">
                        <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                      </div>
                    )}
                  </div>

                  <div className="flex-1 space-y-3">
                    <h3 className="text-sm font-semibold text-gray-900 mb-2">Como configurar:</h3>
                    <ol className="text-sm text-gray-600 space-y-2 list-none">
                      <li className="flex items-start gap-2">
                        <span className="flex-shrink-0 w-5 h-5 rounded-full bg-navy text-white text-xs font-bold flex items-center justify-center mt-0.5">1</span>
                        Abra o <span className="font-medium text-gray-900 mx-1">Google Authenticator</span> ou <span className="font-medium text-gray-900 mx-1">Authy</span>
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="flex-shrink-0 w-5 h-5 rounded-full bg-navy text-white text-xs font-bold flex items-center justify-center mt-0.5">2</span>
                        Escaneie o QR code ao lado
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="flex-shrink-0 w-5 h-5 rounded-full bg-navy text-white text-xs font-bold flex items-center justify-center mt-0.5">3</span>
                        Digite o código de 6 dígitos gerado abaixo
                      </li>
                    </ol>
                  </div>
                </div>

                <div className="space-y-3">
                  <label className="block text-sm font-semibold text-gray-700">Código de verificação</label>
                  <input
                    type="text"
                    inputMode="numeric"
                    maxLength={6}
                    value={confirmCode}
                    onChange={e => setConfirmCode(e.target.value.replace(/\D/g, ''))}
                    placeholder="000 000"
                    className="nxd-input w-44 text-center text-2xl font-mono tracking-widest"
                  />
                  <p className="text-xs text-gray-400">O código muda a cada 30 segundos</p>
                </div>

                <div className="flex items-center gap-3">
                  <button
                    onClick={handleConfirm2FA}
                    disabled={actionLoading || confirmCode.length !== 6}
                    className="nxd-btn nxd-btn-primary"
                  >
                    {actionLoading ? <Loader2 className="w-5 h-5 animate-spin" /> : <CheckCircle className="w-5 h-5" />}
                    {actionLoading ? 'Verificando...' : 'Confirmar e Ativar'}
                  </button>
                  <button
                    onClick={cancelSetup}
                    className="text-sm text-gray-500 hover:text-gray-700 px-4 py-2.5 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    Cancelar
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {isModalOpen && <RegenerateKeyModal onClose={() => setIsModalOpen(false)} />}
    </>
  );
}
