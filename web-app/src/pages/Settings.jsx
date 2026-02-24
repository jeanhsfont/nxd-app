import React, { useState, useEffect } from 'react';
import api from '../utils/api';
import RegenerateKeyModal from '../components/RegenerateKeyModal';

// Icons as inline SVG components
const KeyIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="7.5" cy="15.5" r="5.5"/><path d="m21 2-9.6 9.6"/><path d="m15.5 7.5 3 3L22 7l-3-3"/>
  </svg>
);

const ShieldIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
  </svg>
);

const CheckIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="20 6 9 17 4 12"/>
  </svg>
);

const LoaderIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{animation: 'spin 1s linear infinite'}}>
    <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
  </svg>
);

const XIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
  </svg>
);

export default function Settings() {
  const [isModalOpen, setIsModalOpen] = useState(false);

  // 2FA state
  const [twoFaEnabled, setTwoFaEnabled] = useState(false);
  const [twoFaLoading, setTwoFaLoading] = useState(true);
  const [showSetup, setShowSetup] = useState(false);
  const [qrCodeUrl, setQrCodeUrl] = useState(null);
  const [confirmCode, setConfirmCode] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [showDisableInput, setShowDisableInput] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [message, setMessage] = useState(null); // { type: 'success'|'error', text: string }

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
      <style>{`
        @keyframes spin { to { transform: rotate(360deg); } }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
        .fade-in { animation: fadeIn 0.3s ease forwards; }
      `}</style>

      <div className="p-8 max-w-2xl">
        <h1 className="text-3xl font-bold text-gray-900 mb-1">Configurações</h1>
        <p className="text-gray-500 mb-8 text-sm">Gerencie sua chave de API e segurança da conta.</p>

        {/* Global message */}
        {message && (
          <div className={`fade-in mb-6 flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium border ${
            message.type === 'success'
              ? 'bg-emerald-50 text-emerald-800 border-emerald-200'
              : 'bg-red-50 text-red-800 border-red-200'
          }`}>
            {message.type === 'success' ? <CheckIcon /> : <XIcon />}
            {message.text}
          </div>
        )}

        {/* API Key Card */}
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm mb-5 overflow-hidden">
          <div className="px-6 py-5 border-b border-gray-100 flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-indigo-50 flex items-center justify-center text-indigo-600">
              <KeyIcon />
            </div>
            <div>
              <h2 className="text-base font-semibold text-gray-900">Chave de API</h2>
              <p className="text-xs text-gray-500">Para conectar o DX Simulator e dispositivos externos</p>
            </div>
          </div>
          <div className="px-6 py-5">
            <p className="text-sm text-gray-600 mb-5 leading-relaxed">
              Ao gerar uma nova chave, sua chave atual será <span className="font-medium text-gray-800">invalidada permanentemente</span>.
              A nova chave só será exibida uma única vez por segurança.
            </p>
            <button
              onClick={() => setIsModalOpen(true)}
              className="inline-flex items-center gap-2 bg-red-50 hover:bg-red-100 text-red-700 font-semibold text-sm py-2.5 px-4 rounded-xl border border-red-200 transition-colors"
            >
              <KeyIcon />
              Gerar Nova Chave de API
            </button>
          </div>
        </div>

        {/* 2FA Card */}
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
          <div className="px-6 py-5 border-b border-gray-100 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={`w-9 h-9 rounded-xl flex items-center justify-center ${twoFaEnabled ? 'bg-emerald-50 text-emerald-600' : 'bg-gray-50 text-gray-500'}`}>
                <ShieldIcon />
              </div>
              <div>
                <h2 className="text-base font-semibold text-gray-900">Autenticação de Dois Fatores</h2>
                <p className="text-xs text-gray-500">TOTP — Google Authenticator, Authy, etc.</p>
              </div>
            </div>
            {!twoFaLoading && (
              <span className={`inline-flex items-center gap-1.5 text-xs font-semibold px-3 py-1.5 rounded-full ${
                twoFaEnabled
                  ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                  : 'bg-gray-100 text-gray-500 border border-gray-200'
              }`}>
                {twoFaEnabled ? <><CheckIcon /> Ativo</> : 'Inativo'}
              </span>
            )}
          </div>

          <div className="px-6 py-5">
            {twoFaLoading ? (
              <div className="flex items-center gap-2 text-gray-400 text-sm py-2">
                <LoaderIcon /> Verificando status...
              </div>
            ) : twoFaEnabled ? (
              /* --- 2FA ENABLED STATE --- */
              <div className="fade-in">
                <div className="flex items-start gap-3 p-4 bg-emerald-50 rounded-xl border border-emerald-100 mb-5">
                  <div className="text-emerald-500 mt-0.5"><CheckIcon /></div>
                  <div>
                    <p className="text-sm font-semibold text-emerald-800">Conta protegida com 2FA</p>
                    <p className="text-xs text-emerald-600 mt-0.5">Cada acesso exigirá confirmação pelo app autenticador.</p>
                  </div>
                </div>

                {!showDisableInput ? (
                  <button
                    onClick={() => { setShowDisableInput(true); setMessage(null); }}
                    className="inline-flex items-center gap-2 text-sm text-gray-500 hover:text-red-600 font-medium py-2 px-3 rounded-lg hover:bg-red-50 transition-colors border border-transparent hover:border-red-100"
                  >
                    <XIcon /> Desativar 2FA
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
                        className="w-36 px-4 py-2.5 text-center text-lg font-mono tracking-widest border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                      />
                      <button
                        onClick={handleDisable2FA}
                        disabled={actionLoading}
                        className="flex items-center gap-2 bg-red-600 hover:bg-red-700 disabled:bg-red-300 text-white text-sm font-semibold py-2.5 px-4 rounded-xl transition-colors"
                      >
                        {actionLoading ? <LoaderIcon /> : <XIcon />}
                        Desativar
                      </button>
                      <button
                        onClick={() => { setShowDisableInput(false); setDisableCode(''); setMessage(null); }}
                        className="text-sm text-gray-500 hover:text-gray-700 px-3 py-2.5 rounded-xl hover:bg-gray-100 transition-colors"
                      >
                        Cancelar
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ) : !showSetup ? (
              /* --- 2FA DISABLED STATE --- */
              <div className="fade-in">
                <p className="text-sm text-gray-600 mb-5 leading-relaxed">
                  Adicione uma camada extra de segurança à sua conta. Após ativar, você precisará de um
                  código temporário do app autenticador para fazer login.
                </p>
                <button
                  onClick={handleSetup2FA}
                  disabled={actionLoading}
                  className="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-300 text-white font-semibold text-sm py-2.5 px-5 rounded-xl transition-colors"
                >
                  {actionLoading ? <LoaderIcon /> : <ShieldIcon />}
                  {actionLoading ? 'Gerando QR Code...' : 'Ativar Autenticação de Dois Fatores'}
                </button>
              </div>
            ) : (
              /* --- 2FA SETUP FLOW --- */
              <div className="fade-in space-y-6">
                <div className="flex items-start gap-4">
                  {/* QR Code */}
                  <div className="flex-shrink-0">
                    {qrCodeUrl ? (
                      <div className="w-44 h-44 rounded-2xl border-2 border-indigo-100 bg-white p-2 shadow-sm">
                        <img src={qrCodeUrl} alt="QR Code 2FA" className="w-full h-full rounded-xl" />
                      </div>
                    ) : (
                      <div className="w-44 h-44 rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50 flex items-center justify-center">
                        <LoaderIcon />
                      </div>
                    )}
                  </div>

                  {/* Instructions */}
                  <div className="flex-1 space-y-3">
                    <div>
                      <h3 className="text-sm font-semibold text-gray-900 mb-1">Como configurar:</h3>
                      <ol className="text-sm text-gray-600 space-y-1.5 list-none">
                        <li className="flex items-start gap-2">
                          <span className="flex-shrink-0 w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 text-xs font-bold flex items-center justify-center mt-0.5">1</span>
                          Abra o <span className="font-medium text-gray-800 mx-1">Google Authenticator</span> ou <span className="font-medium text-gray-800 mx-1">Authy</span>
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="flex-shrink-0 w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 text-xs font-bold flex items-center justify-center mt-0.5">2</span>
                          Escaneie o QR code ao lado
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="flex-shrink-0 w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 text-xs font-bold flex items-center justify-center mt-0.5">3</span>
                          Digite o código de 6 dígitos gerado abaixo
                        </li>
                      </ol>
                    </div>
                  </div>
                </div>

                {/* Code input */}
                <div className="space-y-3">
                  <label className="block text-sm font-semibold text-gray-700">Código de verificação</label>
                  <input
                    type="text"
                    inputMode="numeric"
                    maxLength={6}
                    value={confirmCode}
                    onChange={e => setConfirmCode(e.target.value.replace(/\D/g, ''))}
                    placeholder="000 000"
                    className="w-44 px-4 py-3 text-center text-2xl font-mono tracking-widest border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                  <p className="text-xs text-gray-400">O código muda a cada 30 segundos</p>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-3 pt-1">
                  <button
                    onClick={handleConfirm2FA}
                    disabled={actionLoading || confirmCode.length !== 6}
                    className="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-200 disabled:cursor-not-allowed text-white font-semibold text-sm py-2.5 px-5 rounded-xl transition-colors"
                  >
                    {actionLoading ? <LoaderIcon /> : <CheckIcon />}
                    {actionLoading ? 'Verificando...' : 'Confirmar e Ativar'}
                  </button>
                  <button
                    onClick={cancelSetup}
                    className="text-sm text-gray-500 hover:text-gray-700 px-4 py-2.5 rounded-xl hover:bg-gray-100 transition-colors"
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
