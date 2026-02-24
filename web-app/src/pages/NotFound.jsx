import React, { useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { FileQuestion } from 'lucide-react';

/**
 * Página 404 profissional. Layout consistente: se autenticado oferece "Voltar ao Dashboard";
 * se não, "Ir para Login". Sem tela branca, sem comportamento indefinido.
 */
export default function NotFound() {
  const hasToken = typeof window !== 'undefined' && !!localStorage.getItem('nxd-token');
  const location = useLocation();

  useEffect(() => {
    document.title = 'Página não encontrada | NXD';
    return () => { document.title = 'NXD'; };
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center px-4 font-sans">
      <div className="text-center max-w-md">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-indigo-100 border border-indigo-200 mb-6">
          <FileQuestion className="w-8 h-8 text-indigo-600" aria-hidden="true" />
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Página não encontrada</h1>
        <p className="text-gray-600 text-sm mb-6">
          A rota <code className="bg-gray-200 px-1.5 py-0.5 rounded text-xs">{location.pathname}</code> não existe ou foi movida.
        </p>
        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          {hasToken ? (
            <Link
              to="/"
              className="inline-flex items-center justify-center px-5 py-2.5 bg-indigo-600 text-white font-semibold rounded-lg hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
            >
              Voltar ao Dashboard
            </Link>
          ) : (
            <Link
              to="/login"
              className="inline-flex items-center justify-center px-5 py-2.5 bg-indigo-600 text-white font-semibold rounded-lg hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
            >
              Ir para Login
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}
