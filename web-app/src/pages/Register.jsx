import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

export default function Register() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [termsAccepted, setTermsAccepted] = useState(false);
  const [error, setError] = useState('');
  const [termsTouched, setTermsTouched] = useState(false);
  const navigate = useNavigate();

  const termsError = termsTouched && !termsAccepted;
  const canSubmit = termsAccepted;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    if (!termsAccepted) {
      setTermsTouched(true);
      setError('É necessário aceitar os Termos de Uso para criar a conta.');
      return;
    }

    try {
      const response = await fetch('/api/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, email, password }),
      });

      if (!response.ok) {
        let msg = 'Falha ao registrar';
        try {
          const errData = await response.json();
          msg = errData.error || msg;
        } catch {
          msg = response.status >= 500 ? 'Sistema temporariamente indisponível. Tente em instantes.' : msg;
        }
        throw new Error(msg);
      }

      navigate('/login');
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50">
      <div className="p-8 bg-white rounded-xl shadow-md border border-gray-200 w-full max-w-sm">
        <h1 className="text-2xl font-bold mb-4 text-center text-gray-900">Criar Conta</h1>
        <form onSubmit={handleSubmit} noValidate>
          <div className="mb-4">
            <label htmlFor="register-name" className="block text-gray-700 font-semibold mb-2">
              Nome
            </label>
            <input
              type="text"
              id="register-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent ${error ? 'border-red-500' : 'border-gray-300'}`}
              required
              aria-invalid={!!error}
              aria-describedby={error ? 'register-error' : undefined}
            />
          </div>
          <div className="mb-4">
            <label htmlFor="register-email" className="block text-gray-700 font-semibold mb-2">
              Email
            </label>
            <input
              type="email"
              id="register-email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent ${error ? 'border-red-500' : 'border-gray-300'}`}
              required
              aria-invalid={!!error}
              aria-describedby={error ? 'register-error' : undefined}
            />
          </div>
          <div className="mb-4">
            <label htmlFor="register-password" className="block text-gray-700 font-semibold mb-2">
              Senha
            </label>
            <input
              type="password"
              id="register-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent ${error ? 'border-red-500' : 'border-gray-300'}`}
              required
              aria-invalid={!!error}
              aria-describedby={error ? 'register-error' : undefined}
            />
          </div>
          <div className="mb-4">
            <div className="flex items-start gap-3">
              <input
                type="checkbox"
                id="register-terms"
                checked={termsAccepted}
                onChange={(e) => {
                  setTermsAccepted(e.target.checked);
                  setTermsTouched(true);
                }}
                className="mt-1 h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                aria-describedby={termsError ? 'register-terms-error' : undefined}
                aria-invalid={termsError || undefined}
              />
              <div className="text-sm text-gray-700 flex-1">
                <label htmlFor="register-terms" className="cursor-pointer">
                  Li e aceito os
                </label>
                {' '}
                <Link to="/terms" target="_blank" rel="noopener noreferrer" className="text-indigo-600 hover:underline font-medium">
                  Termos de Uso
                </Link>
              </div>
            </div>
            {termsError && (
              <p id="register-terms-error" className="text-red-600 text-sm mt-1" role="alert" aria-live="assertive">
                É necessário aceitar os Termos de Uso para criar a conta.
              </p>
            )}
          </div>
          {error && (
            <p id="register-error" className="text-red-600 text-sm mb-4 p-2 bg-red-50 border border-red-200 rounded-lg" role="alert">
              {error}
            </p>
          )}
          <button
            type="submit"
            disabled={!canSubmit}
            className="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2.5 px-4 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
          >
            Registrar
          </button>
        </form>
        <p className="text-center text-sm text-gray-600 mt-4">
          Já tem uma conta?{' '}
          <Link to="/login" className="text-indigo-600 hover:underline font-medium">
            Faça o login
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
