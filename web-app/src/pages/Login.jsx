import { useState, useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';

export default function Login() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [showSessionToast, setShowSessionToast] = useState(false);

  useEffect(() => {
    const reason = searchParams.get('reason');
    const fromStorage = typeof sessionStorage !== 'undefined' && sessionStorage.getItem('nxd_session_expired_toast');
    if (reason === 'session_expired' || fromStorage) {
      setError('Sessão expirada. Faça login novamente.');
      setShowSessionToast(true);
      setSearchParams({}, { replace: true });
      try {
        sessionStorage.removeItem('nxd_session_expired_toast');
      } catch (_) {}
    }
  }, [searchParams, setSearchParams]);

  useEffect(() => {
    if (!showSessionToast) return;
    const t = setTimeout(() => setShowSessionToast(false), 5000);
    return () => clearTimeout(t);
  }, [showSessionToast]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        let msg = 'Credenciais inválidas';
        try {
          const errorData = await response.json();
          msg = errorData.error || msg;
        } catch {
          msg = response.status >= 500 ? 'Sistema temporariamente indisponível. Tente em instantes.' : msg;
        }
        throw new Error(msg);
      }

      const { token } = await response.json();
      localStorage.setItem('nxd-token', token);

      // Redireciona para o dashboard ou outra página protegida
      window.location.href = '/onboarding';
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50 relative">
      {showSessionToast && (
        <div
          className="absolute top-4 left-1/2 -translate-x-1/2 z-10 px-4 py-2 bg-amber-600 text-white text-sm font-medium rounded-lg shadow-lg"
          role="alert"
          aria-live="polite"
        >
          Sessão expirada
        </div>
      )}
      <div className="p-8 bg-white rounded-xl shadow-md border border-gray-200 w-full max-w-sm">
        <h1 className="text-2xl font-bold mb-4 text-center text-gray-900">NXD Login</h1>
        <form onSubmit={handleSubmit} noValidate>
          <div className="mb-4">
            <label htmlFor="login-email" className="block text-gray-700 font-semibold mb-2">
              Email
            </label>
            <input
              type="email"
              id="login-email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent ${error ? 'border-red-500' : 'border-gray-300'}`}
              required
              aria-invalid={!!error}
              aria-describedby={error ? 'login-error' : undefined}
            />
          </div>
          <div className="mb-6">
            <label htmlFor="login-password" className="block text-gray-700 font-semibold mb-2">
              Senha
            </label>
            <input
              type="password"
              id="login-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent ${error ? 'border-red-500' : 'border-gray-300'}`}
              required
              aria-invalid={!!error}
              aria-describedby={error ? 'login-error' : undefined}
            />
          </div>
          {error && (
            <p id="login-error" className="text-red-600 text-sm mb-4 p-2 bg-red-50 border border-red-200 rounded-lg" role="alert">
              {error}
            </p>
          )}
          <button
            type="submit"
            className="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2.5 px-4 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
          >
            Entrar
          </button>
        </form>
        <p className="text-center text-sm text-gray-600 mt-4">
          Não tem uma conta?{' '}
          <Link to="/register" className="text-indigo-600 hover:underline font-medium">
            Crie uma agora
          </Link>
        </p>
        <p className="text-center text-xs text-gray-500 mt-3">
          <Link to="/terms" className="text-indigo-600 hover:underline">Termos de Uso</Link>
          {' · '}
          <Link to="/support" className="text-indigo-600 hover:underline">Suporte</Link>
        </p>
      </div>
    </div>
  );
}

