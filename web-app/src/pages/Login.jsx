import { useState, useEffect } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { Mail, Lock, Eye, EyeOff, ArrowRight, Zap } from 'lucide-react';
import toast from 'react-hot-toast';
import api from '../utils/api';

export default function Login() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState('email'); // 'email' | '2fa'
  const [tempToken, setTempToken] = useState('');
  const [code2fa, setCode2fa] = useState('');

  useEffect(() => {
    if (searchParams.get('reason') === 'session_expired') {
      toast.error('Sessão expirada. Faça login novamente.');
      setSearchParams({}, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (step === '2fa') {
      if (!code2fa.trim()) {
        toast.error('Digite o código do app autenticador');
        return;
      }
      setLoading(true);
      try {
        const response = await api.post('/api/login/2fa', { code: code2fa.trim() }, {
          headers: { Authorization: `Bearer ${tempToken}` },
        });
        const token = response.data.token;
        localStorage.setItem('nxd-token', token);
        toast.success('Login realizado!');
        setTimeout(() => navigate('/'), 300);
      } catch (err) {
        const msg = err.response?.data?.error || err.message;
        toast.error(msg);
      } finally {
        setLoading(false);
      }
      return;
    }

    if (!email || !password) {
      toast.error('Preencha todos os campos');
      return;
    }

    setLoading(true);
    try {
      const response = await api.post('/api/login', { email, password });
      if (response.data.requires_2fa && response.data.temp_token) {
        setTempToken(response.data.temp_token);
        setStep('2fa');
        toast.success('Digite o código do app autenticador');
      } else {
        const token = response.data.token;
        localStorage.setItem('nxd-token', token);
        toast.success('Login realizado!');
        setTimeout(() => navigate('/'), 300);
      }
    } catch (err) {
      const msg = err.response?.data?.error || err.message;
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 bg-navy rounded-lg mb-4">
            <Zap className="w-7 h-7 text-white" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-1">Bem-vindo ao NXD</h1>
          <p className="text-gray-500 text-sm">Faça login para continuar</p>
        </div>

        {/* Card */}
        <div className="nxd-card">
          {step === '2fa' ? (
            <form onSubmit={handleSubmit} className="space-y-5">
              <p className="text-sm text-gray-600">Digite o código de 6 dígitos do seu app autenticador.</p>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Código 2FA</label>
                <input
                  type="text"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  maxLength={6}
                  value={code2fa}
                  onChange={(e) => setCode2fa(e.target.value.replace(/\D/g, ''))}
                  className="nxd-input text-center text-lg tracking-widest"
                  placeholder="000000"
                />
              </div>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => { setStep('email'); setCode2fa(''); setTempToken(''); }}
                  className="nxd-btn flex-1 justify-center bg-gray-200 text-gray-800 hover:bg-gray-300"
                >
                  Voltar
                </button>
                <button
                  type="submit"
                  disabled={loading || code2fa.length < 6}
                  className="nxd-btn nxd-btn-primary flex-1 justify-center"
                >
                  {loading ? <div className="spinner" /> : <>Confirmar <ArrowRight className="w-5 h-5" /></>}
                </button>
              </div>
            </form>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-5">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Email</label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="nxd-input pl-10"
                    placeholder="seu@email.com"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Senha</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="nxd-input pl-10 pr-10"
                    placeholder="••••••••"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  >
                    {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="nxd-btn nxd-btn-primary w-full justify-center"
              >
                {loading ? (
                  <div className="spinner"></div>
                ) : (
                  <>
                    Entrar
                    <ArrowRight className="w-5 h-5" />
                  </>
                )}
              </button>
            </form>
          )}

          <div className="mt-6 text-center">
            <p className="text-gray-600 text-sm">
              Não tem uma conta?{' '}
              <Link to="/register" className="text-navy hover:underline font-medium">
                Crie uma agora
              </Link>
            </p>
          </div>

          <div className="mt-4 flex items-center justify-center gap-4 text-xs text-gray-500">
            <Link to="/terms" className="hover:text-gray-700">Termos de Uso</Link>
            <span>·</span>
            <Link to="/support" className="hover:text-gray-700">Suporte</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
